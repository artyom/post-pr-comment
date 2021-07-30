# Comment on a GitHub pull request

This GitHub actions adds a comment on a pull request, using provided template and variables to fill it.

## Inputs

* `github-token` (**required**): GitHub token to access API to post comment. Usually this is the `${{ github.token }}` variable.

  If you limit the default [permission scope](https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#permissions) for the token, make sure to grant `pull-requests: write` access.
* `template-file` (**required**): path to the template file. Template is processed using Go [text/template package](https://pkg.go.dev/text/template), its documentation provides details on supported syntax.
* `variables-file` (**required**): path to the JSON file with the template variables mapping.

## Outputs

* `comment-id`: integer ID of the comment created.
* `comment-url`: full URL pointing to the comment created.

## Example usage

```yaml
steps:
  - uses: actions/checkout@v2
  - uses: artyom/post-pr-comment@v1
    with:
      github-token: ${{ github.token }}
      template-file: template.txt
      variables-file: vars.json
```

Alternatively, you [reference use a published docker image][uses-doc] directly:

```yaml
steps:
  - uses: actions/checkout@v2
  - uses: docker://ghcr.io/artyom/post-pr-comment:v1
    with:
      github-token: ${{ github.token }}
      template-file: template.txt
      variables-file: vars.json
```

### Template example

Template file like this:

```text
Here's your list of {{.name}}:

{{range .items -}}
* {{.}}
{{end}}
```

And a variable mapping JSON file for this template:

```json
{
 "name": "animals",
 "items": [
  "dog",
  "cat"
 ]
}
```

Produce such an output:

```text
Here's your list of animals:

* dog
* cat
```

You can experiment with rendering in the playground: <https://play.golang.org/p/TOl7HrQHnVT>

[uses-doc]: https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idstepsuses
