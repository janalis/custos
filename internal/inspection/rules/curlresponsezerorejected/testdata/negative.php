<?php
curl_setopt($ch, CURLOPT_RETURNTRANSFER, true); if (curl_exec($ch) === false) { throw new RuntimeException(); }
