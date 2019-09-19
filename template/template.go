package template

// HTMLContent is the base template used for gowitness reports
var HTMLContent = `
<!doctype html>
<html lang="en">

<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
  <meta name="description" content="">
  <meta name="author" content="Leon Jacobs @leonjza">
  <link rel="icon" href="favicon.ico">

  <title>gowitness - report</title>

  <!-- Bootstrap core CSS -->
  <style> ` + bootstrapCSS + ` </style>

  <!-- Custom styles for this template -->
  <style>
    body {
      min-height: 75rem;
      /* Can be removed; just added for demo purposes */
    }

    .navbar {
      margin-bottom: 0;
    }

    .jumbotron {
      padding-top: 6rem;
      padding-bottom: 6rem;
      margin-bottom: 0;
      background-color: #fff;
    }

    .jumbotron p:last-child {
      margin-bottom: 0;
    }

    .jumbotron-heading {
      font-weight: 300;
    }

    .jumbotron .container {
      max-width: 40rem;
    }

    .album {
      padding-top: 3rem;
      padding-bottom: 3rem;
      background-color: #f7f7f7;
    }

    .card {
      float: left;
      padding: .75rem;
      margin-bottom: 2rem;
      border: 0;
    }

    .card>img {
      margin-bottom: .75rem;
    }

    .card-text {
      font-size: 85%;
    }

    footer {
      padding-top: 3rem;
      padding-bottom: 3rem;
    }

    footer p {
      margin-bottom: .25rem;
    }
  </style>
</head>

<body>

  <header>
    <div class="collapse bg-dark" id="navbarHeader">
      <div class="container">
        <div class="row">
          <div class="col-sm-8 py-4">
            <h4 class="text-white">About</h4>
            <p class="text-muted">This report contains all of the screenshots taken during the gowitness session.</p>
          </div>
          <div class="col-sm-4 py-4">
            <h4 class="text-white">Contact</h4>
            <ul class="list-unstyled">
              <li> <a href="https://github.com/sensepost/gowitness" class="text-white">Github</a> </li>
              <li> <a href="https://twitter.com/leonjza" class="text-white">@leonjza</a> </li>
              <li> <a href="https://twitter.com/sensepost" class="text-white">@sensepost</a> </li>
            </ul>
          </div>
        </div>
      </div>
    </div>
    <div class="navbar navbar-dark bg-dark">
      <div class="container d-flex justify-content-between">
        <a href="#" class="navbar-brand">gowitness report</a>
        <button class="navbar-toggler" type="button" data-toggle="collapse" data-target="#navbarHeader" aria-controls="navbarHeader"
          aria-expanded="false" aria-label="Toggle navigation">
          <span class="navbar-toggler-icon"></span>
        </button>
      </div>
    </div>
  </header>

  <main role="main">

    <section class="jumbotron text-center">
      <div class="container">
        <h4 class="jumbotron-heading">This gowitness report page contains {{ len .ScreenShots }} screenshot(s)!</h4>
      </div>
    </section>

    <div class="album text-muted">

      {{ $report_name := .ReportName }}
      {{ $current_page := .CurrentPage }}

      <nav aria-label="Page navigation example">
        <ul class="pagination justify-content-center">
          {{ range $i, $p := .Pages }}
            <li class="page-item {{ if eq $i $current_page }}active{{ end }}">
              <a class="page-link" href="{{ printf "%s-%d.html" $report_name $i }}">
                {{ $i }}
              </a>
            </li>
          {{ end }}
        </ul>
      </nav>

      <div class="container">

        {{ range $screenshot := .ScreenShots }}

        <div class="row">

          <section>
            <div class="container py-3">
              <div class="card">
                <div class="row ">
                  <div class="col-md-4">
                    <a href="{{ $screenshot.ScreenshotFile }}" target="_blank" rel="noopener noreferrer">
                      <img src="{{ $screenshot.ScreenshotFile }}" class="w-100">
                    </a>
                  </div>
                  <div class="col-md-8 px-3">
                    <div class="card-block px-3">
                      <h4 class="card-title">
                        <a href="{{ $screenshot.FinalURL }}" target="_blank" rel="noopener noreferrer">{{ $screenshot.URL}}</a>
                        <small>{{ $screenshot.ResponseCodeString }}</small>
                        <br>
                        <small class="text-muted">Title: {{ $screenshot.Title }}</small>
                      </h4>
                      <p class="card-text">

                        <!-- headers -->
                        <table class="table table-sm">
                          <thead>
                            <tr>
                              <th scope="col">Header</th>
                              <th scope="col">Value</th>
                            </tr>
                          </thead>
                          <tbody>

                            {{ range $header := $screenshot.Headers }}
                            <tr>
                              <td>{{ $header.Key }}</td>
                              <td>
                                <span class="d-inline-block text-truncate" style="max-width: 450px;">
                                  {{ $header.Value }}
                                </span>
                              </td>
                            </tr>
                            {{ end }}

                          </tbody>
                        </table>

                        <!-- ssl -->
                        <p class="h6">SSL DNS Names: </p>
                        <ul>
                          {{ range $certificates := $screenshot.SSL.PeerCertificates }} {{ range $dnsname := $certificates.DNSNames }}
                          <li>
                            {{ $dnsname }}
                          </li>
                          {{ end }} {{ end }}
                        </ul>

                      </p>
                      <p class="card-text"></p>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </section>

        </div>

        {{ end }}

      </div>

      <nav aria-label="Page navigation example">
        <ul class="pagination justify-content-center">
          {{ range $i, $p := .Pages }}
            <li class="page-item {{ if eq $i $current_page }}active{{ end }}">
              <a class="page-link" href="{{ printf "%s-%d.html" $report_name $i }}">
                {{ $i }}
              </a>
            </li>
          {{ end }}
        </ul>
      </nav>

    </div>

  </main>

  <footer class="text-muted">
    <div class="container">
      <p class="float-right">
        <a href="#">Back to top</a>
      </p>
    </div>
  </footer>

  <script>` + jquery + `</script>
  <script>` + unveil + `</script>
  <script>` + popper + `</script>
  <script>` + bootstrap + `</script>
</body>

</html>
`

// PlaceHolderImage is the image used when a screenshot is missing
var PlaceHolderImage = `
data:image/gif;base64,iVBORw0KGgoAAAANSUhEUgAAAoAAAAHgCAYAAAA10dzkAAAAAXNSR0IArs4c6QAAAAlwSFlzAAALEwAACxMBAJqcGAAAA6ppVFh0WE1MOmNvbS5hZG9iZS54bXAAAAAAADx4OnhtcG1ldGEgeG1sbnM6eD0iYWRvYmU6bnM6bWV0YS8iIHg6eG1wdGs9IlhNUCBDb3JlIDUuNC4wIj4KICAgPHJkZjpSREYgeG1sbnM6cmRmPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5LzAyLzIyLXJkZi1zeW50YXgtbnMjIj4KICAgICAgPHJkZjpEZXNjcmlwdGlvbiByZGY6YWJvdXQ9IiIKICAgICAgICAgICAgeG1sbnM6eG1wPSJodHRwOi8vbnMuYWRvYmUuY29tL3hhcC8xLjAvIgogICAgICAgICAgICB4bWxuczp0aWZmPSJodHRwOi8vbnMuYWRvYmUuY29tL3RpZmYvMS4wLyIKICAgICAgICAgICAgeG1sbnM6ZXhpZj0iaHR0cDovL25zLmFkb2JlLmNvbS9leGlmLzEuMC8iPgogICAgICAgICA8eG1wOk1vZGlmeURhdGU+MjAxNy0xMS0xMVQxMjoxMTo3NDwveG1wOk1vZGlmeURhdGU+CiAgICAgICAgIDx4bXA6Q3JlYXRvclRvb2w+UGl4ZWxtYXRvciAzLjQuMTwveG1wOkNyZWF0b3JUb29sPgogICAgICAgICA8dGlmZjpPcmllbnRhdGlvbj4xPC90aWZmOk9yaWVudGF0aW9uPgogICAgICAgICA8dGlmZjpDb21wcmVzc2lvbj41PC90aWZmOkNvbXByZXNzaW9uPgogICAgICAgICA8dGlmZjpSZXNvbHV0aW9uVW5pdD4yPC90aWZmOlJlc29sdXRpb25Vbml0PgogICAgICAgICA8dGlmZjpZUmVzb2x1dGlvbj43MjwvdGlmZjpZUmVzb2x1dGlvbj4KICAgICAgICAgPHRpZmY6WFJlc29sdXRpb24+NzI8L3RpZmY6WFJlc29sdXRpb24+CiAgICAgICAgIDxleGlmOlBpeGVsWERpbWVuc2lvbj42NDA8L2V4aWY6UGl4ZWxYRGltZW5zaW9uPgogICAgICAgICA8ZXhpZjpDb2xvclNwYWNlPjE8L2V4aWY6Q29sb3JTcGFjZT4KICAgICAgICAgPGV4aWY6UGl4ZWxZRGltZW5zaW9uPjQ4MDwvZXhpZjpQaXhlbFlEaW1lbnNpb24+CiAgICAgIDwvcmRmOkRlc2NyaXB0aW9uPgogICA8L3JkZjpSREY+CjwveDp4bXBtZXRhPgpS3JvOAABAAElEQVR4Ae3dB5gsVZ034LqYc86YFRURXEFRxCzigphFBETWHGBXRcA14ZoTiAEQcwDTKuCa1gRmQUVUEMWcA2YwoSgfv/qs3uqqmanuudNzp899z/P07e7q01V13tNz+98n1bott9zy39etW/e482+bVRIBAgQIECBAgECxAuedd943z78dtm6rrbY6Q/BXbD0rGAECBAgQIEBgTCBB4CaCvzETTwgQIECAAAECRQsk9tuk6BIqHAECBAgQIECAQE9AANgjsYEAAQIECBAgULaAALDs+lU6AgQIECBAgEBPQADYI7GBAAECBAgQIFC2gACw7PpVOgIECBAgQIBAT0AA2COxgQABAgQIECBQtoAAsOz6VToCBAgQIECAQE9AANgjsYEAAQIECBAgULaAALDs+lU6AgQIECBAgEBPQADYI7GBAAECBAgQIFC2gACw7PpVOgIECBAgQIBAT0AA2COxgQABAgQIECBQtoAAsOz6VToCBAgQIECAQE9AANgjsYEAAQIECBAgULaAALDs+lU6AgQIECBAgEBPQADYI7GBAAECBAgQIFC2gACw7PpVOgIECBAgQIBAT0AA2COxgQABAgQIECBQtoAAsOz6VToCBAgQIECAQE9AANgjsYEAAQIECBAgULaAALDs+lU6AgQIECBAgEBPQADYI7GBAAECBAgQIFC2gACw7PpVOgIECBAgQIBAT0AA2COxgQABAgQIECBQtoAAsOz6VToCBAgQIECAQE9AANgjsYEAAQIECBAgULaAALDs+lU6AgQIECBAgEBPQADYI7GBAAECBAgQIFC2gACw7PpVOgIECBAgQIBAT0AA2COxgQABAgQIECBQtoAAsOz6VToCBAgQIECAQE9AANgjsYEAAQIECBAgULaAALDs+lU6AgQIECBAgEBPQADYI7GBAAECBAgQIFC2gACw7PpVOgIECBAgQIBAT0AA2COxgQABAgQIECBQtoAAsOz6VToCBAgQIECAQE9AANgjsYEAAQIECBAgULaAALDs+lU6AgQIECBAgEBPQADYI7GBAAECBAgQIFC2gACw7PpVOgIECBAgQIBAT0AA2COxgQABAgQIECBQtoAAsOz6VToCBAgQIECAQE9AANgjsYEAAQIECBAgULaAALDs+lU6AgQIECBAgEBPQADYI7GBAAECBAgQIFC2gACw7PpVOgIECBAgQIBAT0AA2COxgQABAgQIECBQtoAAsOz6VToCBAgQIECAQE9AANgjsYEAAQIECBAgULaAALDs+lU6AgQIECBAgEBPQADYI7GBAAECBAgQIFC2gACw7PpVOgIECBAgQIBAT0AA2COxgQABAgQIECBQtoAAsOz6VToCBAgQIECAQE9AANgjsYEAAQIECBAgULaAALDs+lU6AgQIECBAgEBPQADYI7GBAAECBAgQIFC2gACw7PpVOgIECBAgQIBAT0AA2COxgQABAgQIECBQtoAAsOz6VToCBAgQIECAQE9AANgjsYEAAQIECBAgULaAALDs+lU6AgQIECBAgEBPQADYI7GBAAECBAgQIFC2gACw7PpVOgIECBAgQIBAT0AA2COxgQABAgQIECBQtoAAsOz6VToCBAgQIECAQE9AANgjsYEAAQIECBAgULaAALDs+lU6AgQIECBAgEBPQADYI7GBAAECBAgQIFC2gACw7PpVOgIECBAgQIBAT0AA2COxgQABAgQIECBQtoAAsOz6VToCBAgQIECAQE9AANgjsYEAAQIECBAgULaAALDs+lU6AgQIECBAgEBPQADYI7GBAAECBAgQIFC2gACw7PpVOgIECBAgQIBAT0AA2COxgQABAgQIECBQtoAAsOz6VToCBAgQIECAQE9AANgjsYEAAQIECBAgULaAALDs+lU6AgQIECBAgEBPQADYI7GBAAECBAgQIFC2gACw7PpVOgIECBAgQIBAT0AA2COxgQABAgQIECBQtoAAsOz6VToCBAgQIECAQE9AANgjsYEAAQIECBAgULaAALDs+lU6AgQIECBAgEBPQADYI7GBAAECBAgQIFC2gACw7PpVOgIECBAgQIBAT0AA2COxgQABAgQIECBQtoAAsOz6VToCBAgQIECAQE9AANgjsYEAAQIECBAgULaAALDs+lU6AgQIECBAgEBPQADYI7GBAAECBAgQIFC2gACw7PpVOgIECBAgQIBAT0AA2COxgQABAgQIECBQtoAAsOz6VToCBAgQIECAQE9AANgjsYEAAQIECBAgULaAALDs+lU6AgQIECBAgEBPQADYI7GBAAECBAgQIFC2gACw7PpVOgIECBAgQIBAT0AA2COxgQABAgQIECBQtoAAsOz6VToCBAgQIECAQE9AANgjsYEAAQIECBAgULaAALDs+lU6AgQIECBAgEBPQADYI7GBAAECBAgQIFC2gACw7PpVOgIECBAgQIBAT0AA2COxgQABAgQIECBQtoAAsOz6VToCBAgQIECAQE9AANgjsYEAAQIECBAgULaAALDs+lU6AgQIECBAgEBPQADYI7GBAAECBAgQIFC2gACw7PpVOgIECBAgQIBAT0AA2COxgQABAgQIECBQtoAAsOz6VToCBAgQIECAQE9AANgjsYEAAQIECBAgULaAALDs+lU6AgQIECBAgEBPQADYI7GBAAECBAgQIFC2gACw7PpVOgIECBAgQIBAT0AA2COxgQABAgQIECBQtoAAsOz6VToCBAgQIECAQE9AANgjsYEAAQIECBAgULaAALDs+lU6AgQIECBAgEBPQADYI7GBAAECBAgQIFC2gACw7PpVOgIECBAgQIBAT0AA2COxgQABAgQIECBQtoAAsOz6VToCBAgQIECAQE9AANgjsYEAAQIECBAgULaAALDs+lU6AgQIECBAgEBPQADYI7GBAAECBAgQIFC2gACw7PpVOgIECBAgQIBAT0AA2COxgQABAgQIECBQtoAAsOz6VToCBAgQIECAQE9AANgjsYEAAQIECBAgULaAALDs+lU6AgQIECBAgEBPQADYI7GBAAECBAgQIFC2gACw7PpVOgIECBAgQIBAT0AA2COxgQABAgQIECBQtoAAsOz6VToCBAgQIECAQE9AANgjsYEAAQIECBAgULaAALDs+lU6AgQIECBAgEBPQADYI7GBAAECBAgQIFC2gACw7PpVOgIECBAgQIBAT0AA2COxgQABAgQIECBQtoAAsOz6VToCBAgQIECAQE9AANgjsYEAAQIECBAgULaAALDs+lU6AgQIECBAgEBPQADYI7GBAAECBAgQIFC2gACw7PpVOgIECBAgQIBAT0AA2COxgQABAgQIECBQtoAAsOz6VToCBAgQIECAQE9AANgjsYEAAQIECBAgULaAALDs+lU6AgQIECBAgEBPQADYI7GBAAECBAgQIFC2gACw7PpVOgIECBAgQIBAT0AA2COxgQABAgQIECBQtoAAsOz6VToCBAgQIECAQE9AANgjsYEAAQIECBAgULaAALDs+lU6AgQIECBAgEBPQADYI7GBAAECBAgQIFC2gACw7PpVOgIECBAgQIBAT0AA2COxgQABAgQIECBQtoAAsOz6VToCBAgQIECAQE9AANgjsYEAAQIECBAgULaAALDs+lU6AgQIECBAgEBPQADYI7GBAAECBAgQIFC2gACw7PpVOgIECBAgQIBAT0AA2COxgQABAgQIECBQtoAAsOz6VToCBAgQIECAQE9AANgjsYEAAQIECBAgULaAALDs+lU6AgQIECBAgEBPQADYI7GBAAECBAgQIFC2gACw7PpVOgIECBAgQIBAT0AA2COxgQABAgQIECBQtoAAsOz6VToCBAgQIECAQE9AANgjsYEAAQIECBAgULaAALDs+lU6AgQIECBAgEBPQADYI7GBAAECBAgQIFC2gACw7PpVOgIECBAgQIBAT0AA2COxgQABAgQIECBQtoAAsOz6VToCBAgQIECAQE9AANgjsYEAAQIECBAgULaAALDs+lU6AgQIECBAgEBPQADYI7GBAAECBAgQIFC2gACw7PpVOgIECBAgQIBAT0AA2COxgQABAgQIECBQtoAAsOz6VToCBAgQIECAQE9AANgjsYEAAQIECBAgULaAALDs+lU6AgQIECBAgEBPQADYI7GBAAECBAgQIFC2gACw7PpVOgIECBAgQIBAT0AA2COxgQABAgQIECBQtoAAsOz6VToCBAgQIECAQE9AANgjsYEAAQIECBAgULaAALDs+h0r3SabbFJd/OIXH9vmCQECBAgQILDxCVxw4yvyxlHiC1/4wtVOO+1U3fSmN62ucY1rVJtuuml11atetbrQhS5UnXXWWdVPfvKT6qc//Wl9+/znP199+tOf3jhglHLNCGyzzTbVBS/4f/8FffnLX67+8pe/rJnzcyIrK7DZZptVl7/85Uc7/e53v1udeeaZo+elPNhYyllKfW3M5Vh385vf/LyNGaC0sl/mMpepdt1112q33XYb+892qJynn3569epXv7r6xCc+MZTV6wRWRODjH/94lc9rk+5973tXP/jBD5qn7gsTeOlLX1rd8Y53HJXqWc96VnXssceOnpfyYGMpZyn1tTGX4/9+fm/MCoWU/YEPfGD1+Mc/vrroRS86dYk2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i+DsCRAgQIAAgcUEMpktE5xyTd8mZQH0zEzPkkiLTXbLpTCzlNMDHvCA5m31fRZCz8oP0nwLmAQy3/Xn7AkQIECAwKBAxgHm6kftS1vmTVkcPVduOvnkk+ur++S675kVn3wJGBM4tlPW/ssVf9qXlWu/7vH8CAgA56eunCkBAgQIEFi2QC7VmaVcttlmm2XtI5cyzFqz6VKW5l9AF/D816ESECBAgACBQYFM/vjQhz5UXwv4mte8ZpWAcJKUtVtz2c8XvvCFlat/TCI2H3m0AM5HPTlLAgQIECCwYgK58sd2221X7bjjjtU1rnGNenH3K13pSvWlRJtF+5v7E088cXQ1kBU7ATva4AICwA1eBU6AAAECBAhseIHM+M0YQGnjELAMzMZRz0pJgAABAgSWFBD8LclT3IsCwOKqVIEIECBAgAABAksLCACX9vEqAQIECBAgQKA4AQFgcVWqQAQIECBAgACBpQUEgEv7eJUAAQIECBAgUJyAALC4KlUgAgQIECBAgMDSAgLApX28SoAAAQIECBAoTkAAWFyVKhABAgQIECBAYGkBAeDSPl4lQIAAAQIECBQnIAAsrkoViAABAgQIECCwtIAAcGkfrxIgQIAAAQIEihMQABZXpQpEgAABAgQIEFha4IJLv+xVAgQIrJ7A9a53vWqLLbaoNt100+qqV71qdc4551Tf+973qve///3V73//+9U7kfOPtNlmm1WXv/zlR8f87ne/W5155pmj5+0H0+Rtv2+tP95yyy2ri1/84qPTPP3006uzzjpr9NwDAgTmV0AAOL9158wJFCOQYG+fffapdtppp2rdunW9cn3rW9+qvvCFL/S2z3LDYx7zmOqOd7zj6BDPetazqmOPPXb0vP1gmrzt9631xwcddFCVoLxJD3/4w6uTTz65eeqeAIE5FhAAznHlOXUCJQhc7GIXq1796ldX17zmNUsojjIQIEBgLgSMAZyLanKSBMoV+Pd///cFg79//OMf1bnnnltuwZWMAAECG1BAC+AGxHdoAgSq6u53v/sYw0knnVS97W1vq0455ZR6vNmVr3zl6s9//vNYHk8IECBAYP0EBIDr5+fdBAish8DlLne56rKXvexoD2n123///auzzz57tG2xiRejDDN68PWvf7268IUvPNr7L37xi9FjDwgQIDDvAgLAea9B509gjgWuda1rjZ39z3/+87Hgb+zFVX6ScYkSAQIEShUwBrDUmlUuAnMgcMELjv8G/ctf/jIHZ+0UCRAgMP8CAsD5r0MlIECAAAECBAhMJTD+83uqt8pMgMCGEMi4tOtc5zrVFa5whXqh4rSa/eY3v6kyRu2nP/3pzE7pAhe4QD1b94pXvGJ93A9/+MMzO9ZydrzJJptUGVN4latcpb5lAeOMJcwC0r/85S9narOc8+2+J8vhXPe6162udKUrVZe85CWrX/3qV6M6/etf/9rNvuznV7/61etFtlOPMctn5zvf+U7161//etn79EYCBOZPQAA4f3XmjDdSge222666973vXd32trcduzpDm+MHP/hBdcIJJ1RHHXXUxF/oCTZe97rXjXZzxhlnVM94xjNGzy9zmctU97vf/apdd921Dqzywt///vdqOQHgjjvuWD30oQ8d7bt9lYlszBVA3vGOd4xez4NPfvKT1WGHHTa2rf3kpje9abX77rtXd73rXccmbbTz5HGu5PGxj32svqpInIbSk570pOqWt7zlKNuRRx5ZHX/88aPnK/Vg2223rR74wAdWqd+LXOQivd1mBvQHP/jB2uWb3/xm7/VJNlziEpeo9txzz+pud7vb2MLOzXvPO++8KpNePvCBD1TvfOc7q7/97W/NS+4JEChUQABYaMUqVjkCCQr222+/6gEPeMBgoa597WtXe++9d3XPe96zevrTn1599rOfHXxPWvZyKbMm/elPf2oeVje84Q2rI444om5tHG1cjwdpoWsfq7urtG52X18s6EkraK7OcbOb3ay7mwWf54oWuT34wQ+uHv/4x1dZbmapdI1rXGPsXBIIr2S60IUuVO277751YLbQ1U+aY6Vl8L73vW99e9Ob3lS9/OUvrzJbetKUMh988MF1q/Fi78nxN9988/qWQP/Zz3529cUvfnGx7LYTIFCAgDGABVSiIpQrcP3rX79uzZsk+Gsr5Bq2r3zlK+tApzvRop1vqccJrNIymK7mtZbSTZrAdNLgr33+F73oResgavvtt29vXtXHmf2cYC7B6FLBX/ekHvKQh1SHH374ki2d7ffssMMO1Vve8pYlg792/jzOub3iFa9Ylm13X54TILB2BbQArt26cWYbucDWW29dd312uwX/8Ic/VJ///Oerr33ta3W3XVrVtthii/oLO/dNSmCRgCFB0iMe8YipWo3SNfui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
`
