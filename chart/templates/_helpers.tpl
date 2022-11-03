{{- define "chart.namespace" -}}
{{- if eq .Release.Namespace "default" -}}
{{- .Values.defaultNamespace -}}
{{- else -}}
{{- .Release.Namespace -}}
{{- end -}}
{{- end -}}