# pdf-Info-app 

## Description  
Returns multiple PDF information in CSV format. The information breakdown is a partial format of the result of the **pdfinfo** command. The following information is returned.  
- File Name
- Author
- Creator
- Producer
- CreationDate
- ModDate
- Page size
- JavaScript
- Pages
- Encrypted
- Page rot
- File size(MB)
- PDF version


## Usage  
```
$ go build main.go
```
Example: rename **main.exe** to **pdf-info-app.exe** and run.
```
$ pdf-Info-app.exe
```

Access http://localhost:14/

## Screen image  
![image](https://user-images.githubusercontent.com/10069642/86309932-d0d4b900-bc57-11ea-8a7a-f63ea82e4ed6.png)  

## Requires  
- Windows
- Go
- nkf
- pdfinfo
- 7zip
