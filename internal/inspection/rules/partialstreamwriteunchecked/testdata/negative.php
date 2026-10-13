<?php
function good($socket, $payload) { return fwrite($socket, $payload) === strlen($payload); }
