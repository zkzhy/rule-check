package taxonomy

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// Mapping represents the IDs for a specific Tactic/Technique/SubTechnique combination
type Mapping struct {
	TacticID         int
	TacticName       string
	TechniqueID      int
	TechniqueName    string
	SubTechniqueID   int
	SubTechniqueName string
	NameEn           string
	CodeOfficial     string
}

type TechniqueCandidate struct {
	TechniqueName string
	SubNames      []string
}

type techniqueNode struct {
	TechniqueID   int
	TechniqueName string
	NameEn        string
	CodeOfficial  string
	Subs          []subNode
	subIdx        map[string]int
}

type subNode struct {
	SubTechniqueID   int
	SubTechniqueName string
	NameEn           string
	CodeOfficial     string
}

var (
	// key: tactic_name|technique_name|sub_technique_name (sub can be empty)
	// value: Mapping
	lookupTable map[string]Mapping
	// key: tactic_name
	// value: tactic_id
	tacticMap map[string]int
	// key: tactic_name
	// value: technique_name -> node
	techByTactic map[string]map[string]*techniqueNode
	once         sync.Once
	loadErr      error
)

// Load initializes the taxonomy from the given CSV path.
// CSV Header expected: tactic_id,tactic_name,technique_id,technique_name,sub_technique_name,sub_technique_id,name_en,code_official
func Load(csvPath string) error {
	once.Do(func() {
		lookupTable = make(map[string]Mapping)
		tacticMap = make(map[string]int)
		techByTactic = make(map[string]map[string]*techniqueNode)
		f, err := os.Open(csvPath)
		if err != nil {
			loadErr = fmt.Errorf("open taxonomy csv failed: %w", err)
			return
		}
		defer f.Close()

		reader := csv.NewReader(f)
		// Skip header
		if _, err := reader.Read(); err != nil {
			loadErr = fmt.Errorf("read header failed: %w", err)
			return
		}

		for {
			record, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				loadErr = fmt.Errorf("read record failed: %w", err)
				return
			}
			if len(record) < 6 {
				continue
			}

			// Parse IDs
			tid, _ := strconv.Atoi(record[0])
			tName := strings.TrimSpace(record[1])
			teid, _ := strconv.Atoi(record[2])
			teName := strings.TrimSpace(record[3])
			subName := strings.TrimSpace(record[4])
			subid, _ := strconv.Atoi(record[5])
			var nameEn, codeOfficial string
			if len(record) > 6 {
				nameEn = strings.TrimSpace(record[6])
			}
			if len(record) > 7 {
				codeOfficial = strings.TrimSpace(record[7])
			}

			m := Mapping{
				TacticID:         tid,
				TacticName:       tName,
				TechniqueID:      teid,
				TechniqueName:    teName,
				SubTechniqueID:   subid,
				SubTechniqueName: subName,
				NameEn:           nameEn,
				CodeOfficial:     codeOfficial,
			}

			// Key format: Tactic|Technique|SubTechnique
			key := makeKey(tName, teName, subName)
			lookupTable[key] = m

			if tid > 0 && tName != "" {
				tacticMap[tName] = tid
			}

			if tName != "" && teName != "" {
				if _, ok := techByTactic[tName]; !ok {
					techByTactic[tName] = make(map[string]*techniqueNode)
				}
				tn, ok := techByTactic[tName][teName]
				if !ok {
					tn = &techniqueNode{
						TechniqueID:   teid,
						TechniqueName: teName,
						subIdx:        make(map[string]int),
					}
					techByTactic[tName][teName] = tn
				}

				if tn.TechniqueID == 0 {
					tn.TechniqueID = teid
				}

				if subName == "" || subid == 0 {
					if tn.NameEn == "" {
						tn.NameEn = nameEn
					}
					if tn.CodeOfficial == "" {
						tn.CodeOfficial = codeOfficial
					}
				} else {
					if _, exists := tn.subIdx[subName]; !exists {
						tn.Subs = append(tn.Subs, subNode{
							SubTechniqueID:   subid,
							SubTechniqueName: subName,
							NameEn:           nameEn,
							CodeOfficial:     codeOfficial,
						})
						tn.subIdx[subName] = len(tn.Subs) - 1
					}
				}
			}
		}
	})
	return loadErr
}

func makeKey(tactic, technique, sub string) string {
	return fmt.Sprintf("%s|%s|%s", strings.TrimSpace(tactic), strings.TrimSpace(technique), strings.TrimSpace(sub))
}

// LookupIDs returns the IDs for the given names.
// If not found, returns (0, 0, 0, false).
func LookupIDs(tactic, technique, sub string) (tacticID, techniqueID, subID int, found bool) {
	if lookupTable == nil {
		return 0, 0, 0, false
	}
	key := makeKey(tactic, technique, sub)
	if m, ok := lookupTable[key]; ok {
		return m.TacticID, m.TechniqueID, m.SubTechniqueID, true
	}
	return 0, 0, 0, false
}

// LookupTacticID returns the ID for a tactic name.
func LookupTacticID(tactic string) (int, bool) {
	if tacticMap == nil {
		return 0, false
	}
	id, ok := tacticMap[strings.TrimSpace(tactic)]
	return id, ok
}

func ListTactics() []string {
	if tacticMap == nil {
		return nil
	}
	type pair struct {
		id   int
		name string
	}
	pairs := make([]pair, 0, len(tacticMap))
	for name, id := range tacticMap {
		pairs = append(pairs, pair{id: id, name: name})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].id != pairs[j].id {
			return pairs[i].id < pairs[j].id
		}
		return pairs[i].name < pairs[j].name
	})
	out := make([]string, 0, len(pairs))
	for _, p := range pairs {
		out = append(out, p.name)
	}
	return out
}

func GenerateTechniqueCandidates(tactic, query string, topK, subMaxPerTechnique int) []TechniqueCandidate {
	tactic = strings.TrimSpace(tactic)
	if techByTactic == nil {
		return nil
	}
	techMap, ok := techByTactic[tactic]
	if !ok || len(techMap) == 0 {
		return nil
	}

	type scored struct {
		name  string
		score int
		id    int
		subs  []string
	}

	queryLower := strings.ToLower(query)
	scoreText := func(needle string, weight int) int {
		needle = strings.TrimSpace(needle)
		if needle == "" {
			return 0
		}
		if strings.Contains(queryLower, strings.ToLower(needle)) || strings.Contains(query, needle) {
			return weight
		}
		return 0
	}

	var scoredList []scored
	for _, tn := range techMap {
		s := scoreText(tn.TechniqueName, 6) + scoreText(tn.NameEn, 4) + scoreText(tn.CodeOfficial, 4)
		var matchedSubs []string
		for _, sub := range tn.Subs {
			subScore := scoreText(sub.SubTechniqueName, 5) + scoreText(sub.NameEn, 3) + scoreText(sub.CodeOfficial, 3)
			if subScore > 0 {
				s += subScore
				matchedSubs = append(matchedSubs, sub.SubTechniqueName)
			}
		}
		scoredList = append(scoredList, scored{
			name:  tn.TechniqueName,
			score: s,
			id:    tn.TechniqueID,
			subs:  matchedSubs,
		})
	}

	sort.Slice(scoredList, func(i, j int) bool {
		if scoredList[i].score != scoredList[j].score {
			return scoredList[i].score > scoredList[j].score
		}
		if scoredList[i].id != scoredList[j].id {
			return scoredList[i].id < scoredList[j].id
		}
		return scoredList[i].name < scoredList[j].name
	})

	anyMatched := false
	for _, it := range scoredList {
		if it.score > 0 {
			anyMatched = true
			break
		}
	}

	var out []TechniqueCandidate
	for _, it := range scoredList {
		if anyMatched && it.score == 0 {
			continue
		}
		tn := techMap[it.name]
		c := TechniqueCandidate{TechniqueName: tn.TechniqueName}

		if len(it.subs) > 0 {
			sort.Strings(it.subs)
			for _, s := range it.subs {
				c.SubNames = append(c.SubNames, s)
				if subMaxPerTechnique > 0 && len(c.SubNames) >= subMaxPerTechnique {
					break
				}
			}
		} else if len(tn.Subs) > 0 {
			sort.Slice(tn.Subs, func(i, j int) bool { return tn.Subs[i].SubTechniqueID < tn.Subs[j].SubTechniqueID })
			for _, sub := range tn.Subs {
				c.SubNames = append(c.SubNames, sub.SubTechniqueName)
				if subMaxPerTechnique > 0 && len(c.SubNames) >= subMaxPerTechnique {
					break
				}
			}
		}

		out = append(out, c)
		if topK > 0 && len(out) >= topK {
			break
		}
	}
	return out
}

func FormatTechniqueCandidates(tactic string, cands []TechniqueCandidate, maxRunes int) string {
	var b strings.Builder
	for _, c := range cands {
		line := c.TechniqueName
		if len(c.SubNames) > 0 {
			line += " (" + strings.Join(c.SubNames, ", ") + ")"
		}
		if b.Len() > 0 {
			b.WriteString("\n")
		}
		b.WriteString("- ")
		b.WriteString(line)

		if maxRunes > 0 && len([]rune(b.String())) >= maxRunes {
			out := []rune(b.String())
			if len(out) > maxRunes {
				return string(out[:maxRunes])
			}
			return string(out)
		}
	}
	return b.String()
}

// GetMapping returns the full Mapping struct
func GetMapping(tactic, technique, sub string) (Mapping, bool) {
	if lookupTable == nil {
		return Mapping{}, false
	}
	key := makeKey(tactic, technique, sub)
	m, ok := lookupTable[key]
	return m, ok
}
