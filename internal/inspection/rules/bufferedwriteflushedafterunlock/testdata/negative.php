<?php
flock($h, LOCK_EX); fwrite($h, $data); fflush($h); flock($h, LOCK_UN);
