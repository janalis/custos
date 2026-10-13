<?php
preg_replace_callback('~[a-z]+~', function($m) { return strtoupper($m[0]); }, $text);
