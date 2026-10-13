<?php
<warning descr="Enable reading HTTP error response bodies.">file_get_contents("https://service.test/data")</warning>;
$ctx = stream_context_create(["http" => ["method" => "GET"]], ["notification" => null]); <warning descr="Enable reading HTTP error response bodies.">file_get_contents("https://service.test/data", false, $ctx)</warning>;
