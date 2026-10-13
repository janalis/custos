<?php
echo 'started'; flush(); <warning descr="Set response headers before committing output.">header('X-State: ready')</warning>;
