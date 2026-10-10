<?php
$h = fopen('data://text/plain,abcdefgh,ij%0A', 'r'); <warning descr="Allow enough bytes for the complete CSV record.">fgetcsv($h, 4)</warning>;
