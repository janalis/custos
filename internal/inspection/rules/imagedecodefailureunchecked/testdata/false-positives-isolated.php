<?php
$im=imagecreatefromstring($bytes); if($im===false){throw new RuntimeException("Invalid image");} imagepng($im);
