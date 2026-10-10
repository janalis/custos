<?php
<warning descr="Remove world-write permission from sensitive files.">chmod("/srv/app/config.php",0777)</warning>;

<warning descr="Remove world-write permission from sensitive files.">mkdir(directory:"/srv/app/.env", permissions:0777)</warning>;
