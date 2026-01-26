{{/*
Common labels
*/}}
{{- define "sales.labels" -}}
app: sales
app.kubernetes.io/name: sales
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version | replace "+" "_" }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "sales.selectorLabels" -}}
app: sales
app.kubernetes.io/name: sales
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Database environment variables for Sales service
*/}}
{{- define "sales.dbEnvVars" -}}
- name: SALES_DB_USER
  valueFrom:
    configMapKeyRef:
      name: sales-config
      key: db_user
- name: SALES_DB_PASSWORD
  valueFrom:
    secretKeyRef:
      name: sales-secret
      key: db_password
- name: SALES_DB_HOST_PORT
  valueFrom:
    configMapKeyRef:
      name: sales-config
      key: db_hostport
- name: SALES_DB_DISABLE_TLS
  valueFrom:
    configMapKeyRef:
      name: sales-config
      key: db_disabletls
{{- end -}}

{{/*
Kubernetes metadata environment variables
*/}}
{{- define "sales.k8sEnvVars" -}}
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
{{- end -}}
