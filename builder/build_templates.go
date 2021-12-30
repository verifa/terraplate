package builder

var tfvarsTemplate = `
{{ range $key, $value := .Variables -}}
{{ $key }}={{ ctyValueToString $value }}
{{- end -}}
`
