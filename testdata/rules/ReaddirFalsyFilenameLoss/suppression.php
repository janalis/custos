<?php
// @custos-ignore ReaddirFalsyFilenameLoss
while ($entry = readdir($dir)) { echo $entry; }
