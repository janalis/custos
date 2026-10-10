<?php
$z=new ZipArchive(); if (<warning descr="Distinguish a missing ZIP entry from index zero.">!$z->locateName("settings.json")</warning>) { echo "missing"; }
