# CurlSslServerSpoofing implementation notes

## Known gaps

- D1 does not check that `K` is the second argument of `curl_setopt`. When
  `K` is the value, it resolves to a number and nothing is reported, but
  `curl_setopt(CURLOPT_SSL_VERIFYPEER, $h, false)` would still be. Requiring
  the second position is equivalent on realistic code and closes that case.
