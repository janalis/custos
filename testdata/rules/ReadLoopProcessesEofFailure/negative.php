<?php
declare(strict_types=1);$h=tmpfile();while(($line=fgets($h))!==false){strlen($line);}while(!feof($h)){$line=fgets($h);if($line!==false){strlen($line);}}while(!feof($h)){echo fgets($h);}while($ready){strlen(fgets($h));}
