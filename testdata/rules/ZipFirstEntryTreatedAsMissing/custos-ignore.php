<?php
// @custos-ignore ZipFirstEntryTreatedAsMissing
$z=new ZipArchive(); if (!$z->locateName("settings.json")) { echo "missing"; }
