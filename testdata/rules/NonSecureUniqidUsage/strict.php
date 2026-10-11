<?php
declare(strict_types=1);

$ids = array_map(<error descr="Pass more_entropy = true to uniqid() to reduce collisions.">'uniqid'</error>, range(1, 3));
$one = call_user_func(<error descr="Pass more_entropy = true to uniqid() to reduce collisions.">'uniqid'</error>, 5);
