<?php
http_response_code(206); <warning descr="Match the body to the inclusive byte range.">header("Content-Range: bytes 0-4/20")</warning>; echo "four"; exit;
