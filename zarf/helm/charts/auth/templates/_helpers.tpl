{{/*
Common labels
*/}}
{{- define "auth.labels" -}}
app: auth
app.kubernetes.io/name: auth
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version | replace "+" "_" }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "auth.selectorLabels" -}}
app: auth
app.kubernetes.io/name: auth
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Database environment variables
*/}}
{{- define "auth.dbEnvVars" -}}
- name: AUTH_DB_USER
  valueFrom:
    configMapKeyRef:
      name: auth-config
      key: db_user
- name: AUTH_DB_PASSWORD
  valueFrom:
    secretKeyRef:
      name: auth-secret
      key: db_password
- name: AUTH_DB_HOST_PORT
  valueFrom:
    configMapKeyRef:
      name: auth-config
      key: db_hostport
- name: AUTH_DB_DISABLE_TLS
  valueFrom:
    configMapKeyRef:
      name: auth-config
      key: db_disabletls
{{- end }}

{{/*
Kubernetes metadata environment variables
*/}}
{{- define "auth.k8sEnvVars" -}}
- name: KUBERNETES_NAMESPACE
  valueFrom:
    fieldRef:
      fieldPath: metadata.namespace
- name: KUBERNETES_NAME
  valueFrom:
    fieldRef:
      fieldPath: metadata.name
- name: KUBERNETES_POD_IP
  valueFrom:
    fieldRef:
      fieldPath: status.podIP
- name: KUBERNETES_NODE_NAME
  valueFrom:
    fieldRef:
      fieldPath: spec.nodeName
{{- end }}
