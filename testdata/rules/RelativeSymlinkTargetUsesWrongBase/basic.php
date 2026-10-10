<?php
chdir('/tmp'); symlink('../data.txt', '/srv/links/current'); <warning descr="Resolve the link target against its directory.">file_get_contents(readlink('/srv/links/current'))</warning>;
