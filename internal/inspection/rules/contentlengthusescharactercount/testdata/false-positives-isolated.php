<?php
function good($body) { header('Content-Length: ' . strlen($body)); echo $body; }
