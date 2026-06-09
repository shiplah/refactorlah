package golang

import (
	"go/token"
	"strings"
	"unicode"

	"github.com/shiplah/refactorlah/internal/adapters/shared"
)

func symbolRenameCandidates(oldBase string, newBase string) []symbolRenameCandidate {
	added := map[string]bool{}
	var candidates []symbolRenameCandidate
	add := func(oldName string, newName string) {
		if !isValidGoIdentifier(oldName) || !isValidGoIdentifier(newName) {
			return
		}
		key := oldName + "\x00" + newName
		if added[key] {
			return
		}
		added[key] = true
		candidates = append(candidates, symbolRenameCandidate{OldName: oldName, NewName: newName})
	}

	add(oldBase, newBase)
	add(goCamelName(oldBase, true, false), goCamelName(newBase, true, false))
	add(goCamelName(oldBase, false, false), goCamelName(newBase, false, false))
	add(goCamelName(oldBase, true, true), goCamelName(newBase, true, true))
	add(goCamelName(oldBase, false, true), goCamelName(newBase, false, true))

	return candidates
}

func goCamelName(value string, exported bool, useInitialisms bool) string {
	parts := splitIdentifierParts(value)
	if len(parts) == 0 {
		return ""
	}

	var builder strings.Builder
	for index, part := range parts {
		lower := strings.ToLower(part)
		if useInitialisms {
			if initialism, ok := goInitialisms[lower]; ok {
				if !exported && index == 0 {
					builder.WriteString(strings.ToLower(initialism))
					continue
				}
				builder.WriteString(initialism)
				continue
			}
		}
		if !exported && index == 0 {
			builder.WriteString(shared.LowerFirst(lower))
			continue
		}
		builder.WriteString(shared.UpperFirst(lower))
	}

	return builder.String()
}

func splitIdentifierParts(value string) []string {
	var parts []string
	var current strings.Builder
	for _, character := range value {
		if unicode.IsLetter(character) || unicode.IsDigit(character) {
			current.WriteRune(character)
			continue
		}
		if current.Len() > 0 {
			parts = append(parts, current.String())
			current.Reset()
		}
	}
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	return parts
}

func isValidGoIdentifier(name string) bool {
	if name == "" || token.Lookup(name).IsKeyword() {
		return false
	}
	for index, character := range name {
		if character == '_' || unicode.IsLetter(character) {
			continue
		}
		if index > 0 && unicode.IsDigit(character) {
			continue
		}
		return false
	}
	return true
}

var goInitialisms = map[string]string{
	"api":   "API",
	"ascii": "ASCII",
	"cpu":   "CPU",
	"css":   "CSS",
	"dns":   "DNS",
	"eof":   "EOF",
	"guid":  "GUID",
	"html":  "HTML",
	"http":  "HTTP",
	"https": "HTTPS",
	"id":    "ID",
	"ip":    "IP",
	"json":  "JSON",
	"qps":   "QPS",
	"ram":   "RAM",
	"rpc":   "RPC",
	"sla":   "SLA",
	"smtp":  "SMTP",
	"sql":   "SQL",
	"ssh":   "SSH",
	"tcp":   "TCP",
	"tls":   "TLS",
	"ttl":   "TTL",
	"udp":   "UDP",
	"ui":    "UI",
	"uid":   "UID",
	"uri":   "URI",
	"url":   "URL",
	"utf8":  "UTF8",
	"uuid":  "UUID",
	"vm":    "VM",
	"xml":   "XML",
}
