package main

/*
#include <jni.h>
#include <stdlib.h>

static const char* get_utf(JNIEnv *env, jstring s) {
    return (*env)->GetStringUTFChars(env, s, NULL);
}

static void release_utf(JNIEnv *env, jstring s, const char* c) {
    (*env)->ReleaseStringUTFChars(env, s, c);
}

static jstring new_utf(JNIEnv *env, const char* c) {
    return (*env)->NewStringUTF(env, c);
}

static void free_cstr(const char* c) {
    free((void*)c);
}
*/
import "C"

import (
	"fmt"
	"os"

	"github.com/Leathal1/hermey-android/core/cachebridge"
)

//export Java_ai_greymattr_hermdroid_core_cache_HermdroidCacheLib_open
func Java_ai_greymattr_hermdroid_core_cache_HermdroidCacheLib_open(env *C.JNIEnv, clazz C.jclass, path C.jstring, maxMessages C.jint, maxBytes C.jlong) C.jint {
	cpath := C.get_utf(env, path)
	defer C.release_utf(env, path, cpath)
	h, err := cachebridge.Open(C.GoString(cpath), int(maxMessages), int64(maxBytes))
	if err != nil {
		fmt.Fprintf(os.Stderr, "cachebridge open error: %v\n", err)
		return C.jint(-1)
	}
	return C.jint(h.Value)
}

//export Java_ai_greymattr_hermdroid_core_cache_HermdroidCacheLib_close
func Java_ai_greymattr_hermdroid_core_cache_HermdroidCacheLib_close(env *C.JNIEnv, clazz C.jclass, handle C.jint) {
	if err := cachebridge.Close(&cachebridge.CacheHandle{Value: int(handle)}); err != nil {
		fmt.Fprintf(os.Stderr, "cachebridge close error: %v\n", err)
	}
}

//export Java_ai_greymattr_hermdroid_core_cache_HermdroidCacheLib_putSession
func Java_ai_greymattr_hermdroid_core_cache_HermdroidCacheLib_putSession(env *C.JNIEnv, clazz C.jclass, handle C.jint, id C.jstring, title C.jstring, lastMessageAtMs C.jlong, count C.jint) C.jint {
	cid := C.GoString(C.get_utf(env, id))
	defer C.release_utf(env, id, C.get_utf(env, id))
	ctitle := C.GoString(C.get_utf(env, title))
	defer C.release_utf(env, title, C.get_utf(env, title))
	err := cachebridge.PutSession(
		&cachebridge.CacheHandle{Value: int(handle)},
		cid, ctitle, int64(lastMessageAtMs), int(count))
	if err != nil {
		fmt.Fprintf(os.Stderr, "cachebridge putSession error: %v\n", err)
		return C.jint(-1)
	}
	return C.jint(0)
}

//export Java_ai_greymattr_hermdroid_core_cache_HermdroidCacheLib_listSessionsJson
func Java_ai_greymattr_hermdroid_core_cache_HermdroidCacheLib_listSessionsJson(env *C.JNIEnv, clazz C.jclass, handle C.jint) C.jstring {
	json, err := cachebridge.ListSessionsJson(&cachebridge.CacheHandle{Value: int(handle)})
	if err != nil {
		fmt.Fprintf(os.Stderr, "cachebridge listSessionsJson error: %v\n", err)
		return 0
	}
	cjson := C.CString(json)
	defer C.free_cstr(cjson)
	return C.new_utf(env, cjson)
}

//export Java_ai_greymattr_hermdroid_core_cache_HermdroidCacheLib_getMessagesJson
func Java_ai_greymattr_hermdroid_core_cache_HermdroidCacheLib_getMessagesJson(env *C.JNIEnv, clazz C.jclass, handle C.jint, sessionID C.jstring) C.jstring {
	csid := C.GoString(C.get_utf(env, sessionID))
	defer C.release_utf(env, sessionID, C.get_utf(env, sessionID))
	json, err := cachebridge.GetMessagesJson(
		&cachebridge.CacheHandle{Value: int(handle)}, csid)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cachebridge getMessagesJson error: %v\n", err)
		return 0
	}
	cjson := C.CString(json)
	defer C.free_cstr(cjson)
	return C.new_utf(env, cjson)
}

//export Java_ai_greymattr_hermdroid_core_cache_HermdroidCacheLib_putMessage
func Java_ai_greymattr_hermdroid_core_cache_HermdroidCacheLib_putMessage(env *C.JNIEnv, clazz C.jclass, handle C.jint, id C.jstring, sessionID C.jstring, role C.jstring, content C.jstring, timestampMs C.jlong) C.jint {
	cid := C.GoString(C.get_utf(env, id))
	defer C.release_utf(env, id, C.get_utf(env, id))
	csid := C.GoString(C.get_utf(env, sessionID))
	defer C.release_utf(env, sessionID, C.get_utf(env, sessionID))
	crole := C.GoString(C.get_utf(env, role))
	defer C.release_utf(env, role, C.get_utf(env, role))
	ccontent := C.GoString(C.get_utf(env, content))
	defer C.release_utf(env, content, C.get_utf(env, content))
	err := cachebridge.PutMessage(
		&cachebridge.CacheHandle{Value: int(handle)},
		cid, csid, crole, ccontent, int64(timestampMs))
	if err != nil {
		fmt.Fprintf(os.Stderr, "cachebridge putMessage error: %v\n", err)
		return C.jint(-1)
	}
	return C.jint(0)
}

//export Java_ai_greymattr_hermdroid_core_cache_HermdroidCacheLib_deleteSession
func Java_ai_greymattr_hermdroid_core_cache_HermdroidCacheLib_deleteSession(env *C.JNIEnv, clazz C.jclass, handle C.jint, sessionID C.jstring) C.jint {
	csid := C.GoString(C.get_utf(env, sessionID))
	defer C.release_utf(env, sessionID, C.get_utf(env, sessionID))
	err := cachebridge.DeleteSession(&cachebridge.CacheHandle{Value: int(handle)}, csid)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cachebridge deleteSession error: %v\n", err)
		return C.jint(-1)
	}
	return C.jint(0)
}

//export Java_ai_greymattr_hermdroid_core_cache_HermdroidCacheLib_evict
func Java_ai_greymattr_hermdroid_core_cache_HermdroidCacheLib_evict(env *C.JNIEnv, clazz C.jclass, handle C.jint) C.jint {
	err := cachebridge.Evict(&cachebridge.CacheHandle{Value: int(handle)})
	if err != nil {
		fmt.Fprintf(os.Stderr, "cachebridge evict error: %v\n", err)
		return C.jint(-1)
	}
	return C.jint(0)
}

func main() {}
