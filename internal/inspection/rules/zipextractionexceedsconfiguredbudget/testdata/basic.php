<?php
$z=new ZipArchive(); $z->open("bundle.zip"); <warning descr="Validate ZIP extraction against configured limits.">$z->extractTo("output")</warning>;
