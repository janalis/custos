<?php
if ($_FILES['a']['type'] === 'image/png') { move_uploaded_file($_FILES['b']['tmp_name'], 'upload'); }
if ($_FILES['a']['type'] === 'image/png') { echo 'rejected'; } else { move_uploaded_file($_FILES['a']['tmp_name'], 'upload'); }
if ($logging) { echo $_FILES['a']['type']; move_uploaded_file($_FILES['a']['tmp_name'], 'upload'); }
