# The body of `make help`. Input: one line per documented target, as the Makefile prints them:
#
#     <name>:<anything>## <description>
#
# The CLASS_* lists arrive as the awk variables `developer` and `production`. A target in
# neither prints under UNCLASSIFIED rather than vanishing, so a missing classification shows
# in the interface itself. `width` is the wrap column: a description wraps under the name
# column instead of being truncated, because its trailing qualifier is what a reader needs.

function assign(list, group,    parts, i, n) {
    n = split(list, parts, " ")
    for (i = 1; i <= n; i++)
        if (parts[i] != "")
            class[parts[i]] = group
}

# `text` wrapped to `width`; the first line follows `lead`, every later one starts at `indent`.
function wrap(lead, text, indent,    words, n, i, line, out) {
    n = split(text, words, " ")
    line = lead
    out = ""
    for (i = 1; i <= n; i++) {
        if (line != lead && line != indent && length(line) + 1 + length(words[i]) > width) {
            out = out line "\n"
            line = indent
        }
        line = line (line == lead || line == indent ? "" : " ") words[i]
    }
    return out line
}

BEGIN {
    order[1] = "developer"
    order[2] = "production"
    order[3] = "UNCLASSIFIED"
    groups = 3

    title["developer"] = "DEVELOPER — the ordinary loop"
    blurb["developer"] = "The gates, the dev servers, the local stack. Safe to run at any moment."
    title["production"] = "PRODUCTION — reaches the deployed stack"
    blurb["production"] = "Prerequisite of nothing. Run one deliberately, and say that you did."
    title["UNCLASSIFIED"] = "UNCLASSIFIED — named by no CLASS_* list in the Makefile"
    blurb["UNCLASSIFIED"] = "Put it in one."

    assign(developer, "developer")
    assign(production, "production")
}

{
    name = substr($0, 1, index($0, ":") - 1)
    where = (name in class) ? class[name] : "UNCLASSIFIED"
    members[where] = members[where] name "\n"
    describes[name] = substr($0, index($0, "## ") + 3)
    if (length(name) > pad)
        pad = length(name)
}

END {
    indent = sprintf("%" (pad + 4) "s", "")
    for (g = 1; g <= groups; g++) {
        group = order[g]
        if (members[group] == "")
            continue
        printf "\n\033[1m%s\033[0m\n", title[group]
        print wrap("  ", blurb[group], "  ")
        print ""
        n = split(members[group], rows, "\n")
        for (i = 1; i <= n; i++) {
            if (rows[i] == "")
                continue
            lead = sprintf("  %-" pad "s  ", rows[i])
            body = wrap(lead, describes[rows[i]], indent)
            coloured = "  \033[36m" rows[i] "\033[0m" substr(lead, 3 + length(rows[i]))
            print coloured substr(body, length(lead) + 1)
        }
    }
    print ""
}
