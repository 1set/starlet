# Documentation coverage check, run by doc_coverage_test.go (dogfoods the regex
# module — starlet checking its own docs with its own tools).
#
# Injected globals:
#   surface: {module_name: [exported member names]} — the authoritative code
#            surface, enumerated in Go from each module's registered members.
#   docs:    {module_name: README text}
#
# A member counts as documented when its name appears immediately after a
# backtick and as a whole identifier — `read_all(`, `ascii_letters`, `I` — which
# is the doc standard's backtick-quoted convention. The next-character guard
# (not an identifier rune) keeps `head` from matching inside `head_lines`.
# RE2 has no lookahead, so the guard character is consumed by the class.
#
# Results read back by the harness: `missing` (list of "module.name") and
# `report` (human summary).
load('regex', 'search', 'escape')

def check():
    missing = []
    documented = 0
    total = 0
    for mod in sorted(surface):
        text = docs[mod]
        for name in surface[mod]:
            total += 1
            if search('`' + escape(name) + '[^A-Za-z0-9_]', text) != None:
                documented += 1
            else:
                missing.append(mod + '.' + name)
    report = str(documented) + '/' + str(total) + ' module members documented across ' + str(len(surface)) + ' modules'
    if missing:
        report = report + '\nUNDOCUMENTED (' + str(len(missing)) + '): ' + ', '.join(sorted(missing))
    return missing, report

missing, report = check()
print(report)
