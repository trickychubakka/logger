package encryption

import (
	"crypto/rsa"
	"log"
	"os"
	"reflect"
	"testing"
)

func TestCreateFile(t *testing.T) {
	type args struct {
		filename string
		data     []byte
	}
	//path, _ := os.Executable()
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Positive CreateFile test",
			args: args{
				filename: "./testfileTEST.test",
				data:     []byte("Test data"),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := CreateFile(tt.args.filename, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("CreateFile() error = %v, wantErr %v", err, tt.wantErr)
			}
			_ = os.Remove(tt.args.filename)
		})
	}
}

func TestDecryptData(t *testing.T) {
	type args struct {
		data       []byte
		privateKey *rsa.PrivateKey
	}

	priv := "./test1234567_id"
	pub := "./test1234567_id.pub"
	privKey, pubKey, err := GenerateRSAKeyPair(priv, pub)
	if err != nil {
		t.Errorf("GenerateRSAKeyPair() error = %v", err)
	}
	defer func(f1, f2 string) {
		log.Println("Remove tmp files")
		err1 := os.Remove(f1)
		err2 := os.Remove(f2)
		if err1 != nil || err2 != nil {
			log.Println("Remove tmp files error err1:", err1, ", err2:", err2)
		}
	}(priv, pub)

	var encryptedData []byte

	encryptedData, err = EncryptData([]byte("Test data"), pubKey)
	if err != nil {
		t.Errorf("EncryptData() error = %v", err)
	}

	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "Positive GenerateRSAKeyPair test",
			args: args{
				data:       encryptedData,
				privateKey: privKey,
			},
			want:    []byte("Test data"),
			wantErr: false,
		},
		{
			name: "Negative GenerateRSAKeyPair test",
			args: args{
				data:       append(encryptedData, 111),
				privateKey: privKey,
			},
			want:    []byte("Test data"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecryptData(tt.args.data, tt.args.privateKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecryptData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DecryptData() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEncryptData(t *testing.T) {
	type args struct {
		data      []byte
		publicKey *rsa.PublicKey
	}

	priv := "./test1234567_id"
	pub := "./test1234567_id.pub"
	_, pubKey, err := GenerateRSAKeyPair(priv, pub)
	if err != nil {
		t.Errorf("GenerateRSAKeyPair() error = %v", err)
	}
	defer func(f1, f2 string) {
		log.Println("Remove tmp files")
		err1 := os.Remove(f1)
		err2 := os.Remove(f2)
		if err1 != nil || err2 != nil {
			log.Println("Remove tmp files error err1:", err1, ", err2:", err2)
		}
	}(priv, pub)

	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "Positive GenerateRSAKeyPair test",
			args: args{
				data:      []byte("Test data"),
				publicKey: pubKey,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EncryptData(tt.args.data, tt.args.publicKey)
			log.Println("encryptedData for 'Test data' is :", string(got))
			if (err != nil) != tt.wantErr {
				t.Errorf("EncryptData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(reflect.TypeOf(got), reflect.TypeOf(tt.want)) {
				t.Errorf("GenerateRSAKeyPair() got1 = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateRSAKeyPair(t *testing.T) {
	type args struct {
		privKeyFile string
		pubKeyFile  string
	}
	tests := []struct {
		name    string
		args    args
		want    *rsa.PrivateKey
		want1   *rsa.PublicKey
		wantErr bool
	}{
		{
			name: "Positive GenerateRSAKeyPair test",
			args: args{
				privKeyFile: "./test1234567_id",
				pubKeyFile:  "./test1234567_id.pub",
			},
			wantErr: false,
		},
		{
			name: "Negative GenerateRSAKeyPair test",
			args: args{
				privKeyFile: "/wrongpath111/test11111/test1234567_id",
				pubKeyFile:  "/wrongpath111/test11111/test1234567_id.pub",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := GenerateRSAKeyPair(tt.args.privKeyFile, tt.args.pubKeyFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateRSAKeyPair() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(reflect.TypeOf(got), reflect.TypeOf(tt.want)) {
				t.Errorf("GenerateRSAKeyPair() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(reflect.TypeOf(got1), reflect.TypeOf(tt.want1)) {
				t.Errorf("GenerateRSAKeyPair() got1 = %v, want %v", got1, tt.want1)
			}
			_ = os.Remove(tt.args.privKeyFile)
			_ = os.Remove(tt.args.pubKeyFile)
		})
	}
}

func TestReadFile(t *testing.T) {
	type args struct {
		filename string
		data     []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Positive ReadFile test",
			args: args{
				filename: "./testfileTEST.test",
				data:     []byte("Test data"),
			},
			wantErr: false,
		},
		{
			name: "Negative ReadFile test",
			args: args{
				filename: "./testfileTEST.test",
				data:     []byte("Test data"),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.wantErr {
				err := CreateFile(tt.args.filename, tt.args.data)
				if err != nil {
					t.Errorf("CreateFile() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
			_, err := ReadFile(tt.args.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				_ = os.Remove(tt.args.filename)
			}
		})
	}
}

func TestReadPrivateKeyFile(t *testing.T) {
	type args struct {
		privKeyFile string
	}

	priv := "./test1234567_id"
	pub := "./test1234567_id.pub"
	_, _, err := GenerateRSAKeyPair(priv, pub)
	if err != nil {
		t.Errorf("GenerateRSAKeyPair() error = %v", err)
	}
	defer func(f1, f2 string) {
		log.Println("Remove tmp files")
		err1 := os.Remove(f1)
		err2 := os.Remove(f2)
		if err1 != nil || err2 != nil {
			log.Println("Remove tmp files error err1:", err1, ", err2:", err2)
		}
	}(priv, pub)

	tests := []struct {
		name    string
		args    args
		want    *rsa.PrivateKey
		wantErr bool
	}{
		{
			name: "Positive ReadPrivateKeyFile test",
			args: args{
				privKeyFile: priv,
			},
			wantErr: false,
		},
		{
			name: "Negative ReadPrivateKeyFile test",
			args: args{
				privKeyFile: "/wrongpath111/test11111/test1234567_id",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadPrivateKeyFile(tt.args.privKeyFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadPrivateKeyFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(reflect.TypeOf(got), reflect.TypeOf(tt.want)) {
				t.Errorf("ReadPrivateKeyFile() got1 = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadPublicKeyFile(t *testing.T) {
	type args struct {
		pubKeyFile string
	}

	priv := "./test1234567_id"
	pub := "./test1234567_id.pub"
	_, _, err := GenerateRSAKeyPair(priv, pub)
	if err != nil {
		t.Errorf("GenerateRSAKeyPair() error = %v", err)
	}
	defer func(f1, f2 string) {
		log.Println("Remove tmp files")
		err1 := os.Remove(f1)
		err2 := os.Remove(f2)
		if err1 != nil || err2 != nil {
			log.Println("Remove tmp files error err1:", err1, ", err2:", err2)
		}
	}(priv, pub)

	tests := []struct {
		name    string
		args    args
		want    *rsa.PublicKey
		wantErr bool
	}{
		{
			name: "Positive ReadPublicKeyFile test",
			args: args{
				pubKeyFile: pub,
			},
			wantErr: false,
		},
		{
			name: "Negative ReadPublicKeyFile test",
			args: args{
				pubKeyFile: "/wrongpath111/test11111/test1234567_id.pub",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadPublicKeyFile(tt.args.pubKeyFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadPublicKeyFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(reflect.TypeOf(got), reflect.TypeOf(tt.want)) {
				t.Errorf("ReadPublicKeyFile() got1 = %v, want %v", got, tt.want)
			}
		})
	}
}
