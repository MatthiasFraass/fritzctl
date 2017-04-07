package completion

const bashCompletionFunctionDefinition string = `
# Pre-defined variables
#
# COMP_LINE : The current command line.
# COMP_WORDS: An array variable consisting of the individual words in the current command line, ${COMP_LINE}.
# COMP_CWORD: An index into ${COMP_WORDS} of the word containing the current cursor position.
# COMPREPLY : An array variable from which bash reads the possible completions generated by a completion function.

_{{.AppName}}()
{
    local current_word prev_word commands_0 flags
    current_word=${COMP_WORDS[COMP_CWORD]}
    prev_word=${COMP_WORDS[COMP_CWORD-1]}
    {{if .Flags}}flags="{{.Flags}}"{{end}}

    case ${COMP_CWORD} in
        {{range $level, $cmdList := .LevelVsCommands}}
        {{$level}})
        {{if eq $level 1}}
            COMPREPLY=($(compgen -W "{{range $index, $cmd := $cmdList}} {{$cmd.Name}}{{end}}" -- ${current_word}))
        {{else}}
        {{end}}
            {{if not $cmdList}}
            {{end}}
	{{end}}
            ;;
        *)
            COMPREPLY=()
            ;;
    esac

}
complete -F _{{.AppName}} {{.AppName}}
`