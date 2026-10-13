<?php
use ZipArchive as BuiltinClass6;
$z=new BuiltinClass6(); $z->open("bundle.zip"); <warning descr="Validate ZIP extraction against configured limits.">$z->extractTo("output")</warning>;
