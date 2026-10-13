<?php
// A registered alias of the BMP type is not a proven mismatch.
header("Content-Type: image/x-ms-bmp"); imagebmp($im);
