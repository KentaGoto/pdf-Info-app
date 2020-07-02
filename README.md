# pdf-Info-app 
BETA  

## Description  
Returns multiple PDF information in CSV format. The information breakdown is a formatted version of the result of the pdfinfo command.  
**Note:**  
Currently, it still only returns some information, such as the application it was created from. Eventually, it will be modified to return other PDF information.  

## Usage  
```
> go run main.go
```

Access http://localhost:12/

## Screen image  
![image](https://user-images.githubusercontent.com/10069642/86309932-d0d4b900-bc57-11ea-8a7a-f63ea82e4ed6.png)  

## Requires  
- Windows
- Go (1.9 or above)
- nkf
- pdfinfo
