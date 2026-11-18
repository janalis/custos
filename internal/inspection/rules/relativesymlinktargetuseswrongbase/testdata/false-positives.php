<?php
chdir('/srv/links'); symlink('../data.txt', '/srv/links/current'); file_get_contents(readlink('/srv/links/current'));
