<?php
echo parse_url('http://example.test/path')['host'];
echo parse_url('https://example.test:65535/')['host'];
echo parse_url('HTTP://example.test:0/')['host'];
echo parse_url('http://[::1]/')['host'];
echo <warning descr="Reject URL parsing failure before indexing.">parse_url('http://example.test:65536/')['host']</warning>;
echo <warning descr="Reject URL parsing failure before indexing.">parse_url('http://example.test:999999999999999999999999/')['host']</warning>;
echo parse_url('http://example.test:%xx/')['host'];
echo <warning descr="Reject URL parsing failure before indexing.">parse_url('http://')['host']</warning>;
echo parse_url('mailto:user@example.test')['host'];
