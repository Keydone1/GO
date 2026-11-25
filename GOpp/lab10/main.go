package main

import (
	"bufio"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strings"
	"sync"
	"time"
)

var in = bufio.NewReader(os.Stdin)

func readLine(prompt string) string {
	fmt.Print(prompt)
	line, _ := in.ReadString('\n')
	return strings.TrimRight(line, "\r\n")
}


// ЗАДАЧА 1. Хэширование (MD5 / SHA-256 / SHA-512) + проверка

func hashWith(algorithm, input string) string {
	switch algorithm {
	case "md5":
		h := md5.Sum([]byte(input))
		return hex.EncodeToString(h[:])
	case "sha256":
		h := sha256.Sum256([]byte(input))
		return hex.EncodeToString(h[:])
	case "sha512":
		h := sha512.Sum512([]byte(input))
		return hex.EncodeToString(h[:])
	}
	return "" // создаем 3 функции хэша
}

func task1() {
	fmt.Println("\nЗадача 1: Хэширование данных")

	text := readLine("Введите строку: ")
	algo := readLine("Алгоритм (md5/sha256/sha512): ")        // выбираем алгоритм
	computed := hashWith(algo, text)
	if computed == "" {
		fmt.Println("❌ Неизвестный алгоритм")
		return
	}
	fmt.Printf("Хэш (%s): %s\n", algo, computed)

	check := readLine("Введите хэш для проверки: ")
	if computed == check {
		fmt.Println("✅ Хэши совпадают. Данные целостны.")
	} else {
		fmt.Println("❌ Хэши не совпадают. Данные повреждены.")
	}
}


// ЗАДАЧА 2. AES-GCM (симметричное шифрование)

func aesEncrypt(plaintext, key string) (string, error) {
	keyHash := sha256.Sum256([]byte(key))
	block, err := aes.NewCipher(keyHash[:])
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(ciphertext), nil                         // создается хэш и сравнивается  при вводе
}

func aesDecrypt(cipherHex, key string) (string, error) {
	keyHash := sha256.Sum256([]byte(key))
	block, err := aes.NewCipher(keyHash[:])
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	data, err := hex.DecodeString(cipherHex)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("данные слишком короткие")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func task2() {
	fmt.Println("\nЗадача 2: AES (симметричное шифрование)")
	text := readLine("Строка для шифрования: ")            // создаем строку и придумываем ключ, далее расшифровка тем же ключом
	key := readLine("Секретный ключ: ")

	encrypted, err := aesEncrypt(text, key)
	if err != nil {
		fmt.Println("❌ Ошибка шифрования:", err)
		return
	}
	fmt.Println("Зашифровано:", encrypted)

	key2 := readLine("Тот же ключ для расшифровки: ")
	decrypted, err := aesDecrypt(encrypted, key2)
	if err != nil {
		fmt.Println("❌ Ошибка расшифровки (неверный ключ?)")
		return
	}
	fmt.Println("Расшифровано:", decrypted)
}


// ЗАДАЧА 3. RSA: подпись и проверка

type rsaKeys struct {
	priv    *rsa.PrivateKey
	privPEM []byte
	pubPEM  []byte
}

func generateRSAKeys() (*rsaKeys, error) {                      //генерация подписи
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	})
	pubDER, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		return nil, err
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})
	return &rsaKeys{priv: priv, privPEM: privPEM, pubPEM: pubPEM}, nil
}

func rsaSign(message string, priv *rsa.PrivateKey) ([]byte, error) {
	hash := sha256.Sum256([]byte(message))
	return rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, hash[:])
}

func rsaVerify(message string, sig []byte, pub *rsa.PublicKey) error {
	hash := sha256.Sum256([]byte(message))
	return rsa.VerifyPKCS1v15(pub, crypto.SHA256, hash[:], sig)
}

func task3() {
	fmt.Println("\n Задача 3: RSA — подпись и проверка ")

	keysA, err := generateRSAKeys()
	if err != nil {
		fmt.Println("❌ Генерация ключей:", err)
		return
	}
	os.WriteFile("private_A.pem", keysA.privPEM, 0600)
	os.WriteFile("public_A.pem", keysA.pubPEM, 0644)
	fmt.Println("✅ Ключи сохранены в private_A.pem, public_A.pem")

	message := "Сообщение 900: перевод 100 000.00 руб."
	fmt.Println("Сообщение:", message)
	sig, _ := rsaSign(message, keysA.priv)
	fmt.Printf("Подпись: %x\n", sig)

	fmt.Println("\n Сторона B проверяет подпись ")
	if rsaVerify(message, sig, &keysA.priv.PublicKey) == nil {
		fmt.Println("✅ Подпись верна")
	} else {
		fmt.Println("❌ Подпись неверна")
	}

	fmt.Println("\n Попытка подделки сообщения")
	if rsaVerify("Перевод 1000000$", sig, &keysA.priv.PublicKey) != nil {
		fmt.Println("✅ Подделка обнаружена")                   //Если сообщения не совпадают с сервером
	}
}

// ЗАДАЧА 4. TLS + mTLS (сертификаты генерируются в памяти)

type certBundle struct {
	cert *x509.Certificate
	key  *rsa.PrivateKey
	der  []byte
	pem  []byte
}

func genCertificate(cn string, isCA bool, parent *certBundle) (*certBundle, error) {           //Генерация сертификата, сервер сам себя обзывает как ему надо
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	serial, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))

	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: cn},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(1 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
		DNSNames:    []string{"localhost"},
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1")},
	}

	parentCert, parentKey := tmpl, key
	if isCA {
		tmpl.IsCA = true
		tmpl.KeyUsage |= x509.KeyUsageCertSign
	} else if parent != nil {
		parentCert, parentKey = parent.cert, parent.key
	}

	der, err := x509.CreateCertificate(rand.Reader, tmpl, parentCert, &key.PublicKey, parentKey)
	if err != nil {
		return nil, err
	}
	cert, _ := x509.ParseCertificate(der)
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	return &certBundle{cert: cert, key: key, der: der, pem: pemBytes}, nil                         //Настройка сертификата, сравнение имени на подлинность
}

func task4() {
	fmt.Println("\n Задача 4: TLS + взаимная аутентификация (mTLS)")

	ca, err := genCertificate("MyCA", true, nil)
	if err != nil {
		fmt.Println("❌ CA:", err)
		return
	}
	server, err := genCertificate("localhost", false, ca)
	if err != nil {
		fmt.Println("❌ Server cert:", err)
		return
	}
	client, err := genCertificate("client", false, ca)
	if err != nil {
		fmt.Println("❌ Client cert:", err)
		return
	}
	fmt.Println("✅ Сертификаты сгенерированы (CA, server, client)")

	caPool := x509.NewCertPool()
	caPool.AppendCertsFromPEM(ca.pem)

	serverCert := tls.Certificate{
		Certificate: [][]byte{server.der},
		PrivateKey:  server.key,
	}
	serverTLS := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    caPool,
	}

	ln, err := tls.Listen("tcp", "127.0.0.1:0", serverTLS)
	if err != nil {
		fmt.Println("❌ Listen:", err)
		return
	}
	addr := ln.Addr().String()
	fmt.Println("TLS-сервер слушает:", addr)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		buf := make([]byte, 1024)
		n, _ := conn.Read(buf)
		fmt.Printf("[Сервер] Получено: %q\n", string(buf[:n]))
		cn := conn.(*tls.Conn).ConnectionState().PeerCertificates[0].Subject.CommonName
		fmt.Printf("[Сервер] Клиент CN=%s (mTLS прошёл)\n", cn)
		io.WriteString(conn, "ACK: сообщение получено\n")
	}()

	time.Sleep(100 * time.Millisecond)

	clientCert := tls.Certificate{
		Certificate: [][]byte{client.der},
		PrivateKey:  client.key,
	}
	clientTLS := &tls.Config{
		RootCAs:      caPool,
		Certificates: []tls.Certificate{clientCert},
		ServerName:   "localhost",
	}

	conn, err := tls.Dial("tcp", addr, clientTLS)
	if err != nil {
		fmt.Println("❌ Клиент:", err)
		ln.Close()
		wg.Wait()
		return
	}
	state := conn.ConnectionState()
	fmt.Printf("[Клиент] Проверил сертификат сервера: CN=%s\n",
		state.PeerCertificates[0].Subject.CommonName)
	io.WriteString(conn, "Секретное сообщение\n")
	reply := make([]byte, 1024)
	n, _ := conn.Read(reply)
	fmt.Printf("[Клиент] Ответ: %q\n", string(reply[:n]))
	conn.Close()
	ln.Close()
	wg.Wait()

	fmt.Println("\n Проверка: без клиентского сертификата mTLS не пропустит ")
	ln2, _ := tls.Listen("tcp", "127.0.0.1:0", serverTLS)
	addr2 := ln2.Addr().String()
	go func() {
		c, _ := ln2.Accept()
		if c != nil {
			c.Close()
		}
	}()
	time.Sleep(50 * time.Millisecond)
	_, err = tls.Dial("tcp", addr2, &tls.Config{
		RootCAs:    caPool,
		ServerName: "localhost",
	})
	if err != nil {
		msg := err.Error()
		if len(msg) > 60 {
			msg = msg[:60] + "..."
		}
		fmt.Println("✅ Сервер отверг клиента без сертификата:", msg)
	}
	ln2.Close()
}

// ЗАДАЧА 5. JWT (вручную, HMAC-SHA256) + роли

var jwtSecret = []byte("super-secret-key-change-me")

func b64url(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }

func signJWT(claims map[string]any) string {                                   //Генерация токена, в котором будет храниться вся информация 
	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	hJSON, _ := json.Marshal(header)
	pJSON, _ := json.Marshal(claims)

	signingInput := b64url(hJSON) + "." + b64url(pJSON)
	mac := hmac.New(sha256.New, jwtSecret)
	mac.Write([]byte(signingInput))
	sig := b64url(mac.Sum(nil))
	return signingInput + "." + sig
}

type JWTClaims struct {
	Sub  string `json:"sub"`
	Role string `json:"role"`
	Exp  int64  `json:"exp"`
}

func verifyJWT(token string) (*JWTClaims, error) {  //Проверка токена на наличие подделки и сертификации
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("неверный формат")
	}
	signingInput := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, jwtSecret)
	mac.Write([]byte(signingInput))
	expected := b64url(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(parts[2])) {
		return nil, fmt.Errorf("неверная подпись")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}
	var c JWTClaims
	if err := json.Unmarshal(payload, &c); err != nil {
		return nil, err
	}
	if c.Exp < time.Now().Unix() {
		return nil, fmt.Errorf("токен истёк")
	}
	return &c, nil
}

func task5() {
	fmt.Println("\n Задача 5: JWT + роли")                       //Показывает вывод ошибок админки и юзеров, а также информацию о сервере

	mux := http.NewServeMux()

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		var creds struct{ Username, Password string }
		json.NewDecoder(r.Body).Decode(&creds)

		role := ""
		switch {
		case creds.Username == "admin" && creds.Password == "admin123":
			role = "admin"
		case creds.Username == "user" && creds.Password == "user123":
			role = "user"
		default:
			http.Error(w, `{"error":"invalid credentials"}`, 401)
			return
		}

		token := signJWT(map[string]any{
			"sub":  creds.Username,
			"role": role,
			"exp":  time.Now().Add(1 * time.Hour).Unix(),
		})
		json.NewEncoder(w).Encode(map[string]string{"token": token})
	})

	auth := func(requiredRole string, next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			if !strings.HasPrefix(h, "Bearer ") {
				http.Error(w, `{"error":"no token"}`, 401)
				return
			}
			claims, err := verifyJWT(strings.TrimPrefix(h, "Bearer "))
			if err != nil {
				http.Error(w, `{"error":"`+err.Error()+`"}`, 401)
				return
			}
			if requiredRole != "" && claims.Role != requiredRole {
				http.Error(w, `{"error":"forbidden"}`, 403)
				return
			}
			next(w, r)
		}
	}

	mux.HandleFunc("/profile", auth("", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"message": "Профиль пользователя"})
	}))
	mux.HandleFunc("/admin", auth("admin", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"message": "Панель администратора"})
	}))

	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	addr := ln.Addr().String()
	srv := &http.Server{Handler: mux}
	go srv.Serve(ln)
	time.Sleep(100 * time.Millisecond)

	baseURL := "http://" + addr
	fmt.Println("JWT-сервер:", baseURL)

	call := func(method, path, token string) {
		req, _ := http.NewRequest(method, baseURL+path, nil)
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println("  ошибка:", err)
			return
		}
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		fmt.Printf("  %s %s -> %d %s\n",
			method, path, resp.StatusCode, strings.TrimSpace(string(b)))
	}

	login := func(user, pass string) string {
		resp, _ := http.Post(baseURL+"/login", "application/json",
			strings.NewReader(`{"username":"`+user+`","password":"`+pass+`"}`))
		var lr struct{ Token string }
		json.NewDecoder(resp.Body).Decode(&lr)
		resp.Body.Close()
		return lr.Token
	}

	adminToken := login("admin", "admin123")
	userToken := login("user", "user123")
	fmt.Println("admin-токен получен:", adminToken[:30], "...")
	fmt.Println("user-токен получен: ", userToken[:30], "...")

	fmt.Println("\n Проверки ")
	call("GET", "/profile", "")
	call("GET", "/profile", "garbage.token.here")
	call("GET", "/profile", userToken)
	call("GET", "/admin", userToken)
	call("GET", "/admin", adminToken)

	srv.Close()
}


// ЗАДАЧА 6. CSRF + сессии

func task6() {
	fmt.Println("\nЗадача 6: CSRF + сессии ")            //подделка запроса с чужого сайта - атака(и защита) от мошеннических сайтов-перехватчиков

	type session struct {
		user string
		csrf string
	}
	var (
		mu       sync.RWMutex
		sessions = map[string]session{}
	)

	randToken := func() string {
		b := make([]byte, 16)
		rand.Read(b)
		return hex.EncodeToString(b)
	}

	mux := http.NewServeMux()

	// GET /login?username=alice → JSON {user, csrf}, ставит cookie session_id
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		username := r.URL.Query().Get("username")
		if username == "" {
			http.Error(w, "username required", 400)
			return
		}
		sid := randToken()
		csrf := randToken()
		mu.Lock()
		sessions[sid] = session{user: username, csrf: csrf}
		mu.Unlock()

		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    sid,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"user": username,
			"csrf": csrf,
		})
	})

	// GET /profile → JSON {user, csrf} по cookie
	mux.HandleFunc("/profile", func(w http.ResponseWriter, r *http.Request) {                 //через  куки файлы происходит валидность жертвы, деньги тырятся успешно
		c, err := r.Cookie("session_id")
		if err != nil {
			http.Error(w, "no session", 401)
			return
		}
		mu.RLock()
		s, ok := sessions[c.Value]
		mu.RUnlock()
		if !ok {
			http.Error(w, "invalid session", 401)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"user": s.user,
			"csrf": s.csrf,
		})
	})

	// POST /transfer — требует X-CSRF-Token
	csrfProtect := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "GET" || r.Method == "HEAD" {
				next(w, r)
				return
			}
			c, err := r.Cookie("session_id")                     //проверка сессии и токенов куки файлов
			if err != nil {
				http.Error(w, "no session", 401)
				return
			}
			mu.RLock()
			s, ok := sessions[c.Value]
			mu.RUnlock()
			if !ok {
				http.Error(w, "invalid session", 401)
				return
			}
			if r.Header.Get("X-CSRF-Token") != s.csrf {
				http.Error(w, "CSRF token invalid", 403)
				return
			}
			next(w, r)
		}
	}

	mux.HandleFunc("/transfer", csrfProtect(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "POST only", 405)
			return
		}
		fmt.Fprintln(w, "✅ Перевод выполнен")
	}))

	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	addr := ln.Addr().String()
	srv := &http.Server{Handler: mux}
	go srv.Serve(ln)
	time.Sleep(100 * time.Millisecond)

	baseURL := "http://" + addr
	fmt.Println("CSRF-сервер:", baseURL)

	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}

	// 1. Логин
	resp, err := client.Get(baseURL + "/login?username=alice")
	if err != nil {
		fmt.Println("❌ login:", err)
		return
	}
	var loginData struct {
		User string `json:"user"`
		CSRF string `json:"csrf"`
	}
	json.NewDecoder(resp.Body).Decode(&loginData)
	resp.Body.Close()
	fmt.Printf("Залогинен как %s, CSRF=%s\n", loginData.User, loginData.CSRF)

	// 2. Проверяем, что cookie сохранилась
	profileResp, _ := client.Get(baseURL + "/profile")
	profileBody, _ := io.ReadAll(profileResp.Body)
	profileResp.Body.Close()
	fmt.Printf("GET /profile -> %d %s",
		profileResp.StatusCode, string(profileBody))

	// 3. POST /transfer с разными CSRF
	doPOST := func(csrfHeader string) {
		req, _ := http.NewRequest("POST", baseURL+"/transfer", nil)
		if csrfHeader != "" {
			req.Header.Set("X-CSRF-Token", csrfHeader)
		}
		r, err := client.Do(req)
		if err != nil {
			fmt.Println("  ошибка:", err)
			return
		}
		defer r.Body.Close()
		b, _ := io.ReadAll(r.Body)
		fmt.Printf("  POST /transfer (csrf=%q) -> %d %s\n",
			csrfHeader, r.StatusCode, strings.TrimSpace(string(b)))
	}

	fmt.Println("\n Проверки ")
	doPOST("")
	doPOST("wrong-token")
	doPOST(loginData.CSRF)

	srv.Close()
}

//main
func main() {
	for {
		fmt.Println("\n Лабораторная 10 ")
		fmt.Println("1. Хэширование (MD5/SHA-256/SHA-512) + проверка")
		fmt.Println("2. AES-шифрование (симметричное)")
		fmt.Println("3. RSA-подпись и проверка")
		fmt.Println("4. TLS + mTLS (взаимная аутентификация)")
		fmt.Println("5. JWT + роли (admin/user)")
		fmt.Println("6. CSRF + сессии")
		fmt.Println("0. Выход")
		fmt.Print("Выбор: ")

		choice := strings.TrimSpace(readLine(""))
		switch choice {
		case "1":
			task1()
		case "2":
			task2()
		case "3":
			task3()
		case "4":
			task4()
		case "5":
			task5()
		case "6":
			task6()
		case "0":
			fmt.Println("Пока!")
			return
		default:
			fmt.Println("❌ Неверный выбор")
		}
	}
}