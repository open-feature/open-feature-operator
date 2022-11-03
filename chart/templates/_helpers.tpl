{{- define "chart.namespace" -}}
    {{- if eq .Release.Namespace default-}}
        {{- .chart.defaultNamespace -}}
    {{- else -}}
        {{- .Release.Namespace -}}
    {{- end -}}
{{- end -}}