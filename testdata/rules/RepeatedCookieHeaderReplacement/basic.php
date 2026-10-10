<?php
header('Set-Cookie: a=1'); <warning descr="Append distinct cookie headers instead of replacing them.">header('Set-Cookie: b=2')</warning>;
