{{/*
Expand the name of the chart.
*/}}
{{- define "account-center.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "account-center.fullname" -}}
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

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "account-center.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "account-center.labels" -}}
helm.sh/chart: {{ include "account-center.chart" . }}
{{ include "account-center.selectorLabels" . }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "account-center.selectorLabels" -}}
app.kubernetes.io/name: {{ include "account-center.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Service account name
*/}}
{{- define "account-center.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "account-center.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Image pull policy
*/}}
{{- define "account-center.imagePullPolicy" -}}
{{- .Values.imagePullPolicy | default "IfNotPresent" }}
{{- end }}

{{/*
Image registry prefix
*/}}
{{- define "account-center.imageRegistry" -}}
{{- if .Values.global.imageRegistry }}
{{- printf "%s/" .Values.global.imageRegistry }}
{{- end }}
{{- end }}
