<?php
echo '<p>' . ($_GET['flag'] === 'yes') . '</p>';
echo '<p>' . (int) $_GET['count'] . '</p>';
echo '<p>' . md5($_GET['label']) . '</p>';
echo '<p>' . hash('sha256', $_GET['label'], binary: false) . '</p>';
