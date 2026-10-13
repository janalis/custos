<?php
$z=new ZipArchive(); <warning descr="Use an existing ZIP entry index.">$z->getNameIndex($z->numFiles)</warning>;
