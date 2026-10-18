<?php
$z=new ZipArchive();$z->open('example.zip',ZipArchive::CREATE);if($z->addFromString('entry','first')){while($flag){$z->addFromString('entry','second');$z->deleteName('entry');}}
