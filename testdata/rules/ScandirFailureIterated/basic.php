<?php
foreach (<warning descr="Check directory enumeration before iterating.">scandir($directory)</warning> as $entry) { echo $entry; }
