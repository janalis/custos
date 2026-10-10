<?php
$z=new ZipArchive(); $z->open("bundle.zip",ZipArchive::CREATE); $z->addFromString("first.txt","one"); $z->addFromString("second.txt","two");
