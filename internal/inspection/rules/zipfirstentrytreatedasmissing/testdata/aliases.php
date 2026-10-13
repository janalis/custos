<?php
use ZipArchive as BuiltinClass6;
$z=new BuiltinClass6(); if (<warning descr="Distinguish a missing ZIP entry from index zero.">!$z->locateName("settings.json")</warning>) { echo "missing"; }
