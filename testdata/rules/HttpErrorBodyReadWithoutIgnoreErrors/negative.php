<?php
$ctx = stream_context_create(["http" => ["ignore_errors" => true]]); file_get_contents("https://service.test/data", false, $ctx);
