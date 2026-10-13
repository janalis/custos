<?php
use ZipArchive as BuiltinClass6;
$z=new BuiltinClass6(); if ($z->locateName("settings.json") === false) { echo "missing"; }
