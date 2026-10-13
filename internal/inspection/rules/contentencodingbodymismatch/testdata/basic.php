<?php
header("Content-Encoding: gzip"); echo <warning descr="Encode the body using the declared content coding.">gzcompress("payload")</warning>; exit;
