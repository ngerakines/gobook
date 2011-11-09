$outlook = New-Object -Com Outlook.Application
$mapi = $outlook.GetNamespace('MAPI')
$mailboxRoot = $mapi.GetDefaultFolder([Microsoft.Office.Interop.Outlook.OlDefaultFolders]::olFolderInbox).Parent
$date = Get-Date
$date

$walkFolderScriptBlock = {
    param(
    	$currentFolder
    )
    foreach ($folder in $currentFolder.Folders) {
		if ($folder.FolderPath.Contains("Sent")) {
			$folder.FolderPath
			foreach ($email in $folder.items) {
				if ($email.SentOn.Day -eq $date.Day -and $email.SentOn.Month -eq $date.Month -and $email.SentOn.Year -eq $date.Year) {
					"#####"
					$people = @()
					if ($email.To -and -not $email.To -eq "") {
						foreach ($person in $email.To.split(";")) {
							$people += "`"@{0}`"" -f $person.Trim()
						}
					}
					if ($email.CC -and -not $email.CC -eq "") {
						foreach ($person in $email.CC.split(";")) {
							$people += "`"@{0}`"" -f $person.Trim()
						}
					}
					if ($email.BCC -and -not $email.BCC -eq "") {
						foreach ($person in $email.BCC.split(";")) {
							$people += "`"@{0}`"" -f $person.Trim()
						}
					}
					foreach ($person in $people) {
						$person
					}
					"`"!{0}-{1}-{2} {3}:{4}`"" -f $email.SentOn.Year, $email.SentOn.Month, $email.SentOn.Day, $email.SentOn.Hour, $email.SentOn.Minute
					$email.Subject
				}
			}
		}
    	& $walkFolderScriptBlock $folder
    }
}
& $walkFolderScriptBlock $mailboxRoot

