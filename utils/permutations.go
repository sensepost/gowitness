package utils

import (
	"math/rand"
	"net/url"
	"strconv"
	"time"
)

// Permutations returns a slice of all of the URL:port combinations
// for a given slice of ips and ports
func Permutations(ips []string, ports []int, skipHTTP bool, skipHTTPS bool) ([]string, error) {

	var results []string

	for _, ip := range ips {

		for _, port := range ports {

			// build a URL
			partialURL := ip + ":" + strconv.Itoa(port)

			// Append an HTTP version
			if !skipHTTP {

				httpURL := HTTP + partialURL

				u, err := url.Parse(httpURL)
				if err != nil {
					return nil, err
				}

				results = append(results, u.String())
			}

			// Append an HTTPS version
			if !skipHTTPS {

				httpsURL := HTTPS + partialURL

				u, err := url.Parse(httpsURL)
				if err != nil {
					return nil, err
				}

				results = append(results, u.String())
			}
		}
	}

	return results, nil
}

// ShufflePermutations and return a new slice.
// 	https://gist.github.com/quux00/8258425
func ShufflePermutations(permutations []string) []string {

	rand.Seed(time.Now().UTC().UnixNano())

	N := len(permutations)
	for i := 0; i < N; i++ {

		r := i + rand.Intn(N-i)
		permutations[r], permutations[i] = permutations[i], permutations[r]
	}

	return permutations
}
