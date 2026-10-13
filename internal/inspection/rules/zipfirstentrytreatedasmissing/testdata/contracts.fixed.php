<?php
$z=new ZipArchive(); if ($z->locateName("settings.json") === false) { echo "missing"; }
