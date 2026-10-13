<?php
http_response_code(206); header("Content-Range: bytes 0-3/20"); echo "four"; exit;
