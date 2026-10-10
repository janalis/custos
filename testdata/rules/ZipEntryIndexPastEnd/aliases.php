<?php
use ZipArchive as BuiltinClass6;
$z=new BuiltinClass6(); <warning descr="Use an existing ZIP entry index.">$z->getNameIndex($z->numFiles)</warning>;
