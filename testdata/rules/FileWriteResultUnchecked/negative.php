<?php
function good($path, $payload) { return file_put_contents($path, $payload) === strlen($payload); }
