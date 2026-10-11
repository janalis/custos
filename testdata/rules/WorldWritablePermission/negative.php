<?php
chmod("/srv/app/config.php",0640);chmod("/srv/app/cache",0777);chmod($unknown,0777);mkdir("/srv/app/.env",$mode);strlen("mode");
chmod('/srv/app/config.php', 0640 | 0004);
chmod('/srv/app/config.php', 0002 | $extra);
chmod('/srv/app/config.php', 0646 & 0775);
mkdir('/srv/app/.env');
