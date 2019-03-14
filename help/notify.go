package help

var EMAIL_NOTIFICATION = Parse(`Hi,
There are some (possible) issues with your backups.
{{ $email := .Email }}
{{ range .Grouped }}
  Rule '{{ .Rule.Name }}'
  Regexp: {{ .Rule.Regexp }}
  {{ range .PerFile }}
    {{ .FullPath }}
    Created: {{ .CreatedAt.Format "Mon Jan 2 15:04:05 MST 2006" }}
    Size: {{ .FileVersion.Size }} b 
    {{ if .Age }}
      Older than expected:
      age: {{ .Age.ActualAge }} 
      overdue: {{ .Age.Overdue }}
    {{ end }}
    {{ if .Size }}
      Weird size delta:
      size {{ .Size.CurrentSize }} b
      previous size {{ .Size.PreviousSize }} b
      delta {{ .Size.Delta }}%
      overflow {{ .Size.Overflow }} b
    {{ end }}
  {{ end }}

  increase expected age +1 day '{{ .Rule.Name }}':
  curl -u {{ $email }} \
    -XPOST \
    -d'{"age_seconds": {{ add .Rule.AcceptableAgeSeconds 86400 }}, "uuid": "{{.Rule.UUID }}"}' \
     https://baxx.dev/protected/change/notification
  
  increase to delta% + 10 for '{{ .Rule.Name }}':
  curl -u {{ $email }} -XPOST \
  -d'{"delta_percent": {{ add .Rule.AcceptableSizeDeltaPercentBetweenVersions 10 }}, "uuid": "{{.Rule.UUID }}"}' \
  https://baxx.dev/protected/change/notification
{{end }}

--
baxx.dev
`)

var EMAIL_AGE_RULE = Parse(`{{ .FullPath }}
  file created: {{ .CreatedAt.Format "Mon Jan 2 15:04:05 MST 2006" }}
  age: {{ .ActualAge }} 
  overdue: {{ .Overdue }}`)

var EMAIL_SIZE_RULE = Parse(`{{ .FullPath }}
  file created: {{ .CreatedAt.Format "Mon Jan 2 15:04:05 MST 2006" }}
  size {{ .CurrentSize }} b
  previous size {{ .PreviousSize }} b
  delta {{ .Delta }}%
  overflow {{ .Overflow }} b`)
