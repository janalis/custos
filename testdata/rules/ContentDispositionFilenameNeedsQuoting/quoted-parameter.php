<?php
header('Content-Disposition: form-data; name="a; filename=b c"');
header('Content-Disposition: form-data; name="a\\"; filename=b c"');
header('Content-Disposition: form-data; name="a; filename=b c');
<warning descr="Quote the Content-Disposition filename.">header('Content-Disposition: form-data; name="a; b"; filename=b c')</warning>;
