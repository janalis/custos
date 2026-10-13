<?php
header("Content-Length: 4"); <warning descr="Send one consistent Content-Length value.">header("Content-Length: 8", false)</warning>;
