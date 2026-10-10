<?php
$h = fopen('data://text/plain,abcdefgh,ij%0A', 'r'); fgetcsv($h, 100);
