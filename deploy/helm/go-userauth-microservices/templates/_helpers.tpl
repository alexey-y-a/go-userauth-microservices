{{- define "go-userauth-microservices.fullname" -}}
{{- printf "%s" .Chart.Name | trunc 63 | trimSuffix "-" -}}
{{- end}}

{{- define "go-userauth-microservices.labels" -}}
app.kubernetes.io/name: {{ include "go-userauth-microservices.fullname" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}