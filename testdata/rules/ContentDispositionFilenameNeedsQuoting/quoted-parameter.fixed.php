<?php
header('Content-Disposition: form-data; name="a; filename=b c"');
header('Content-Disposition: form-data; name="a\\"; filename=b c"');
header('Content-Disposition: form-data; name="a; filename=b c');
header('Content-Disposition: form-data; name="a; b"; filename="b c"');
