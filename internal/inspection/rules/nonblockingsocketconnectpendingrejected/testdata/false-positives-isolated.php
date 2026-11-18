<?php
socket_set_block($s); if (!socket_connect($s, $host, 443)) { throw new RuntimeException(); }
