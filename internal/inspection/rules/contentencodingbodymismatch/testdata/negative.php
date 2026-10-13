<?php
header("Content-Encoding: gzip"); echo gzencode("payload"); exit;
