<?php
flock($h, LOCK_EX); fwrite($h, $data); flock($h, LOCK_UN); <warning descr="Flush buffered writes before releasing the lock.">fflush($h)</warning>;
