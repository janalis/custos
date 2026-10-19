<?php
echo $_GET['prefix'] . ('<p>' . $_GET['label']);
echo '<p>' . htmlspecialchars($_GET['label'], ENT_QUOTES | ENT_SUBSTITUTE, 'UTF-8') . '</p>';
