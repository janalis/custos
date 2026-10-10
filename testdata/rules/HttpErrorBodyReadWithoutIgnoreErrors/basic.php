<?php
$ctx = stream_context_create(["http" => ["ignore_errors" => false]]); <warning descr="Enable reading HTTP error response bodies.">file_get_contents("https://service.test/data", false, $ctx)</warning>;
