{{/*
Expand the name of the chart.
*/}}
{{- define "kubevirt-management.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "kubevirt-management.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{- define "kubevirt-management.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "kubevirt-management.labels" -}}
helm.sh/chart: {{ include "kubevirt-management.chart" . }}
{{ include "kubevirt-management.selectorLabels" . }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- define "kubevirt-management.selectorLabels" -}}
app.kubernetes.io/name: {{ include "kubevirt-management.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "kubevirt-management.serviceAccountName" -}}
{{- default (include "kubevirt-management.fullname" .) .Values.serviceAccount.name }}
{{- end }}

{{- define "kubevirt-management.secretName" -}}
{{- if .Values.auth.existingSecret -}}
{{ .Values.auth.existingSecret }}
{{- else -}}
{{ include "kubevirt-management.fullname" . }}
{{- end -}}
{{- end -}}
