{{- define "gophprofile.name" -}}
{{- default "avatar" .Values.name -}}
{{- end }}

{{- define "gophprofile.serverName" -}}
{{- include "gophprofile.name" . }}-service
{{- end }}

{{- define "gophprofile.workerName" -}}
{{- include "gophprofile.name" . }}-worker
{{- end }}

{{- define "gophprofile.secretName" -}}
{{- if .Values.secret.existingSecret -}}
{{- .Values.secret.existingSecret -}}
{{- else -}}
{{- include "gophprofile.name" . }}-secrets
{{- end -}}
{{- end }}

{{- define "gophprofile.labels" -}}
app.kubernetes.io/name: {{ include "gophprofile.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/part-of: gophprofile
app.kubernetes.io/managed-by: {{ .Release.Service }}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version }}
{{- end }}

{{- define "gophprofile.configChecksums" -}}
checksum/config: {{ include (print .Template.BasePath "/configmap.yaml") . | sha256sum }}
checksum/secret: {{ include (print .Template.BasePath "/secret.yaml") . | sha256sum }}
{{- end }}

{{- define "gophprofile.serverSelectorLabels" -}}
app: {{ include "gophprofile.serverName" . }}
{{- end }}

{{- define "gophprofile.workerSelectorLabels" -}}
app: {{ include "gophprofile.workerName" . }}
{{- end }}

{{- define "gophprofile.metricsIngress" -}}
{{- with .Values.networkPolicy.monitoringNamespaces -}}
- from:
    {{- range . }}
    - namespaceSelector:
        matchLabels:
          kubernetes.io/metadata.name: {{ . }}
    - namespaceSelector:
        matchLabels:
          name: {{ . }}
    {{- end }}
  ports:
    - protocol: TCP
      port: {{ $.Values.ports.metrics }}
{{- end }}
{{- end }}

{{- define "gophprofile.appEgress" -}}
- to:
    - namespaceSelector:
        matchLabels:
          kubernetes.io/metadata.name: kube-system
  ports:
    - protocol: UDP
      port: 53
    - protocol: TCP
      port: 53
{{- if .Values.dependencies.enabled }}
- to:
    - podSelector:
        matchLabels:
          app: postgres
  ports:
    - protocol: TCP
      port: 5432
- to:
    - podSelector:
        matchLabels:
          app: minio
  ports:
    - protocol: TCP
      port: 9000
- to:
    - podSelector:
        matchLabels:
          app: rabbitmq
  ports:
    - protocol: TCP
      port: 5672
{{- end }}
{{- with .Values.networkPolicy.externalEgress }}
{{- toYaml . | nindent 0 }}
{{- end }}
{{- end }}

{{- define "gophprofile.secretEnv" -}}
- name: DATABASE_DSN
  valueFrom:
    secretKeyRef:
      name: {{ include "gophprofile.secretName" . }}
      key: database-dsn
- name: S3_ACCESS_KEY
  valueFrom:
    secretKeyRef:
      name: {{ include "gophprofile.secretName" . }}
      key: s3-access-key
- name: S3_SECRET_KEY
  valueFrom:
    secretKeyRef:
      name: {{ include "gophprofile.secretName" . }}
      key: s3-secret-key
- name: RABBITMQ_URL
  valueFrom:
    secretKeyRef:
      name: {{ include "gophprofile.secretName" . }}
      key: rabbitmq-url
{{- end }}
